use crate::auth::AuthService;
pub mod services;
use self::services::{Character, GameService, LobbyService};
use crate::net::codec::handshake::HandshakeCodec;
use crate::net::codec::world::WorldCodec;
use crate::net::packets::game::GameCommandPacket;
use crate::net::packets::handshake::HandshakePacket;
use crate::net::packets::lobby::LobbyCommandPacket;
use crate::net::packets::world::{WorldCommandPayload, WorldEventPacket};
use futures::{SinkExt, StreamExt};
use std::error::Error;
use std::sync::Arc;
use tokio::net::{TcpListener, TcpStream};
use tokio_util::codec::Framed;
use tracing::{debug, error, info, warn};

pub struct WorldServer {
    addr: String,
    lobby_service: Arc<dyn LobbyService>,
    game_service: Arc<dyn GameService>,
    auth_service: Arc<dyn AuthService>,
}

impl WorldServer {
    pub fn new(
        addr: String,
        lobby_service: Arc<dyn LobbyService>,
        game_service: Arc<dyn GameService>,
        auth_service: Arc<dyn AuthService>,
    ) -> Self {
        Self {
            addr,
            lobby_service,
            game_service,
            auth_service,
        }
    }

    pub async fn run(&self) -> Result<(), Box<dyn Error>> {
        let listener = TcpListener::bind(&self.addr).await?;
        info!("elkia-world listening on: {}", self.addr);

        loop {
            let (socket, addr) = listener.accept().await?;
            info!("Accepted connection from: {}", addr);

            let lobby_service = self.lobby_service.clone();
            let game_service = self.game_service.clone();
            let auth_service = self.auth_service.clone();

            tokio::spawn(async move {
                if let Err(e) =
                    handle_connection(socket, lobby_service, game_service, auth_service).await
                {
                    error!("Error handling connection from {}: {}", addr, e);
                }
            });
        }
    }
}

async fn send_character_list(
    framed: &mut Framed<TcpStream, WorldCodec>,
    chars: &[Character],
) -> Result<(), Box<dyn Error>> {
    framed.send(WorldEventPacket::CharacterListStart(0)).await?;
    for char in chars {
        let packet = format!(
            "c_info {} {} -1 {} {} {} 0 {} {} {} {} {} {} {} {} {} {} 0 0 0 0 0",
            char.name,
            char.id,
            char.class,
            char.level,
            char.hero_level,
            char.hair_color,
            char.hair_style,
            char.faction,
            char.reputation,
            char.dignity,
            char.compliment,
            char.job_level,
            char.experience,
            char.job_experience,
            char.hero_experience
        );
        framed.send(WorldEventPacket::CharacterInfo(packet)).await?;
    }
    framed.send(WorldEventPacket::CharacterListEnd).await?;
    Ok(())
}

async fn handle_connection(
    mut socket: TcpStream,
    lobby_service: Arc<dyn LobbyService>,
    game_service: Arc<dyn GameService>,
    auth_service: Arc<dyn AuthService>,
) -> Result<(), Box<dyn Error>> {
    // 1. Handshake (Session Packet)
    let (sync_cmd, user_cmd, pass_cmd) = {
        let mut framed = Framed::new(&mut socket, HandshakeCodec::new());

        // Read SyncCommand (Packet 1)
        let sync_cmd = match framed.next().await.ok_or("Connection closed (Sync)")?? {
            HandshakePacket::Sync(cmd) => cmd,
            _ => return Err("Expected Sync packet".into()),
        };
        info!("Received SyncCommand: {:?}", sync_cmd);

        // Read UsernameCommand (Packet 2)
        let user_cmd = match framed.next().await.ok_or("Connection closed (User)")?? {
            HandshakePacket::Username(cmd) => cmd,
            _ => return Err("Expected Username packet".into()),
        };
        info!("Received UsernameCommand: {:?}", user_cmd);

        // Read PasswordCommand (Packet 3)
        let pass_cmd = match framed.next().await.ok_or("Connection closed (Pass)")?? {
            HandshakePacket::Password(cmd) => cmd,
            _ => return Err("Expected Password packet".into()),
        };
        info!("Received PasswordCommand: {:?}", pass_cmd);

        (sync_cmd, user_cmd, pass_cmd)
    };

    // Verify Session
    let session = auth_service.verify_handshake(&pass_cmd.password).await?;
    if session.username != user_cmd.username {
        warn!(
            "Security Alert: Username mismatch! Packet: {}, Session: {}",
            user_cmd.username, session.username
        );
        return Err("Username mismatch".into());
    }
    info!(
        "Session verified for user: {} (ID: {})",
        session.username, session.user_id
    );

    // 2. Game Protocol (GatewayCodec)
    let mut framed = Framed::new(socket, WorldCodec::new(sync_cmd.code));

    // Send Character List
    let mut chars = lobby_service.get_characters(&user_cmd.username).await;

    // Create a default character if none exists (for testing)
    if chars.is_empty() {
        if let Ok(_) = lobby_service
            .create_character(&user_cmd.username, "Hero", 1)
            .await
        {
            chars = lobby_service.get_characters(&user_cmd.username).await;
        }
    }

    send_character_list(&mut framed, &chars).await?;

    while let Some(pkt_res) = framed.next().await {
        match pkt_res {
            Ok(cmd) => {
                info!("Received Game Command: {:?}", cmd);
                match cmd.payload {
                    WorldCommandPayload::Heartbeat => {
                        // Respond to heartbeat if needed, usually client sends periodically
                    }
                    WorldCommandPayload::Lobby(lobby_cmd) => match lobby_cmd {
                        LobbyCommandPacket::Select(pkt) => {
                            // Client selects a character slot
                            if pkt.slot < chars.len() {
                                info!("Client selected character: {}", chars[pkt.slot].name);
                                framed.send("OK".to_string()).await?;
                            } else {
                                warn!("Client selected invalid slot: {}", pkt.slot);
                            }
                        }
                        LobbyCommandPacket::GameStart(_) => {
                            info!("Client requested game start");
                            // Send map info, etc.
                            // TODO: Implement map loading logic
                        }
                        LobbyCommandPacket::CharNew(pkt) => {
                            match lobby_service
                                .create_character(&user_cmd.username, &pkt.name, pkt.class)
                                .await
                            {
                                Ok(_) => {
                                    let chars =
                                        lobby_service.get_characters(&user_cmd.username).await;
                                    send_character_list(&mut framed, &chars).await?;
                                }
                                Err(e) => {
                                    error!("Failed to create character: {}", e);
                                }
                            }
                        }
                    },
                    WorldCommandPayload::Game(game_cmd) => match game_cmd {
                        GameCommandPacket::Walk(pkt) => {
                            game_service.walk(&user_cmd.username, pkt.x, pkt.y).await;
                        }
                        GameCommandPacket::Say(pkt) => {
                            game_service.chat(&user_cmd.username, &pkt.message).await;
                        }
                    },
                    WorldCommandPayload::Unknown(tag) => {
                        debug!("Unhandled command: {}", tag);
                    }
                    WorldCommandPayload::Status(_) => {}
                }
            }
            Err(e) => {
                error!("Error reading frame: {}", e);
                break;
            }
        }
    }

    Ok(())
}
