use crate::auth::AuthService;
use crate::net::codec::handshake::HandshakeCodec;
use crate::net::codec::world::WorldCodec;
use crate::net::packets::game::GameCommandPacket;
use crate::net::packets::handshake::HandshakeCommandPacket;
use crate::net::packets::lobby::{
    CharacterInfoPacket, CharacterListEndPacket, CharacterListStartPacket, LobbyCommandPacket,
    LobbyEventPacket, SelectResponsePacket,
};
use crate::net::packets::world::{WorldCommandPayload, WorldEventPacket};
use crate::world::services::{Character, GameService, LobbyService};
use futures::{SinkExt, StreamExt};
use std::error::Error;
use std::sync::Arc;
use tokio::net::TcpStream;
use tokio_util::codec::Framed;
use tracing::{error, info, warn};

pub struct HandshakeState {
    auth_service: Arc<dyn AuthService>,
}

impl HandshakeState {
    pub fn new(auth_service: Arc<dyn AuthService>) -> Self {
        Self { auth_service }
    }

    pub async fn process(
        &self,
        mut socket: TcpStream,
    ) -> Result<(TcpStream, String, u32), Box<dyn Error>> {
        let (sync_cmd, user_cmd, pass_cmd) = {
            let mut framed = Framed::new(&mut socket, HandshakeCodec::new());

            // Read SyncCommand (Packet 1)
            let sync_cmd = match framed.next().await.ok_or("Connection closed (Sync)")?? {
                HandshakeCommandPacket::Sync(cmd) => cmd,
                _ => return Err("Expected Sync packet".into()),
            };
            info!("Received SyncCommand: {:?}", sync_cmd);

            // Read UsernameCommand (Packet 2)
            let user_cmd = match framed.next().await.ok_or("Connection closed (User)")?? {
                HandshakeCommandPacket::Username(cmd) => cmd,
                _ => return Err("Expected Username packet".into()),
            };
            info!("Received UsernameCommand: {:?}", user_cmd);

            // Read PasswordCommand (Packet 3)
            let pass_cmd = match framed.next().await.ok_or("Connection closed (Pass)")?? {
                HandshakeCommandPacket::Password(cmd) => cmd,
                _ => return Err("Expected Password packet".into()),
            };
            info!("Received PasswordCommand: {:?}", pass_cmd);

            (sync_cmd, user_cmd, pass_cmd)
        };

        // Verify Session
        let session = self.auth_service.verify_handshake(&pass_cmd.password).await?;
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

        Ok((socket, user_cmd.username, sync_cmd.code))
    }
}

pub struct LobbyState {
    lobby_service: Arc<dyn LobbyService>,
}

impl LobbyState {
    pub fn new(lobby_service: Arc<dyn LobbyService>) -> Self {
        Self { lobby_service }
    }

    async fn send_character_list(
        &self,
        framed: &mut Framed<TcpStream, WorldCodec>,
        chars: &[Character],
    ) -> Result<(), Box<dyn Error>> {
        framed
            .send(WorldEventPacket::Lobby(
                LobbyEventPacket::CharacterListStart(CharacterListStartPacket { sequence: 0 }),
            ))
            .await?;
        for char in chars {
            framed
                .send(WorldEventPacket::Lobby(LobbyEventPacket::CharacterInfo(
                    CharacterInfoPacket {
                        name: char.name.clone(),
                        id: char.id.clone(),
                        class: char.class,
                        level: char.level,
                        hero_level: char.hero_level,
                        hair_color: char.hair_color,
                        hair_style: char.hair_style,
                        faction: char.faction,
                        reputation: char.reputation,
                        dignity: char.dignity,
                        compliment: char.compliment,
                        job_level: char.job_level,
                        experience: char.experience,
                        job_experience: char.job_experience,
                        hero_experience: char.hero_experience,
                    },
                )))
                .await?;
        }
        framed
            .send(WorldEventPacket::Lobby(
                LobbyEventPacket::CharacterListEnd(CharacterListEndPacket),
            ))
            .await?;
        Ok(())
    }

    pub async fn process(
        &self,
        framed: &mut Framed<TcpStream, WorldCodec>,
        username: &str,
    ) -> Result<Character, Box<dyn Error>> {
        // Send Character List
        let mut chars = self.lobby_service.get_characters(username).await;

        // Create a default character if none exists (for testing)
        if chars.is_empty() {
            if let Ok(_) = self
                .lobby_service
                .create_character(username, "Hero", 1)
                .await
            {
                chars = self.lobby_service.get_characters(username).await;
            }
        }

        self.send_character_list(framed, &chars).await?;

        let mut selected_slot: Option<usize> = None;

        while let Some(pkt_res) = framed.next().await {
            match pkt_res {
                Ok(cmd) => {
                    info!("Lobby Received Command: {:?}", cmd);
                    match cmd.payload {
                        WorldCommandPayload::Heartbeat => {
                            // Heartbeat
                        }
                        WorldCommandPayload::Lobby(lobby_cmd) => match lobby_cmd {
                            LobbyCommandPacket::Select(pkt) => {
                                if pkt.slot < chars.len() {
                                    info!("Client selected character: {}", chars[pkt.slot].name);
                                    selected_slot = Some(pkt.slot);
                                    framed
                                        .send(WorldEventPacket::Lobby(
                                            LobbyEventPacket::SelectResponse(SelectResponsePacket),
                                        ))
                                        .await?;
                                } else {
                                    warn!("Client selected invalid slot: {}", pkt.slot);
                                }
                            }
                            LobbyCommandPacket::GameStart(_) => {
                                if let Some(slot) = selected_slot {
                                    info!("Client requested game start with slot {}", slot);
                                    if slot < chars.len() {
                                        return Ok(chars[slot].clone());
                                    }
                                } else {
                                    warn!("Client requested game start without selection");
                                }
                            }
                            LobbyCommandPacket::CharNew(pkt) => {
                                match self
                                    .lobby_service
                                    .create_character(username, &pkt.name, pkt.class)
                                    .await
                                {
                                    Ok(_) => {
                                        chars = self.lobby_service.get_characters(username).await;
                                        self.send_character_list(framed, &chars).await?;
                                    }
                                    Err(e) => {
                                        error!("Failed to create character: {}", e);
                                    }
                                }
                            }
                        },
                        WorldCommandPayload::Game(_) => {
                            warn!("Received game command in lobby state, ignoring");
                        }
                    }
                }
                Err(e) => {
                    return Err(format!("Error reading frame: {}", e).into());
                }
            }
        }

        Err("Connection closed during lobby".into())
    }
}

pub struct GameState {
    game_service: Arc<dyn GameService>,
}

impl GameState {
    pub fn new(game_service: Arc<dyn GameService>) -> Self {
        Self { game_service }
    }

    pub async fn process(
        &self,
        framed: &mut Framed<TcpStream, WorldCodec>,
        username: &str,
        _character: Character,
    ) -> Result<(), Box<dyn Error>> {
        info!("Entered Game State for user: {}", username);

        // TODO: Implement map loading logic here (send map info, etc.)

        while let Some(pkt_res) = framed.next().await {
            match pkt_res {
                Ok(cmd) => {
                    info!("Game Received Command: {:?}", cmd);
                    match cmd.payload {
                        WorldCommandPayload::Heartbeat => {
                            // Heartbeat
                        }
                        WorldCommandPayload::Lobby(_) => {
                             warn!("Received lobby command in game state, ignoring");
                        }
                        WorldCommandPayload::Game(game_cmd) => match game_cmd {
                            GameCommandPacket::Walk(pkt) => {
                                self.game_service.walk(username, pkt.x, pkt.y).await;
                            }
                            GameCommandPacket::Say(pkt) => {
                                self.game_service.chat(username, &pkt.message).await;
                            }
                        },
                    }
                }
                Err(e) => {
                    return Err(format!("Error reading frame: {}", e).into());
                }
            }
        }

        Ok(())
    }
}
