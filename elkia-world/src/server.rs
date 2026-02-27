use std::error::Error;
use std::sync::Arc;
use std::str::FromStr;
use tokio::net::{TcpListener, TcpStream};
use tokio_util::codec::Framed;
use futures::{SinkExt, StreamExt};
use elkia_net::codec::session::SessionCodec;
use elkia_net::codec::gateway::GatewayCodec;
use elkia_net::packets::session::{SyncCommand, UsernameCommand, PasswordCommand};
use elkia_net::packets::world::{WorldCommandPacket, WorldCommandPayload};
use crate::services::{Character, LobbyService, GameService};

pub struct WorldServer {
    addr: String,
    lobby_service: Arc<dyn LobbyService>,
    game_service: Arc<dyn GameService>,
}

impl WorldServer {
    pub fn new(addr: String, lobby_service: Arc<dyn LobbyService>, game_service: Arc<dyn GameService>) -> Self {
        Self {
            addr,
            lobby_service,
            game_service,
        }
    }

    pub async fn run(&self) -> Result<(), Box<dyn Error>> {
        let listener = TcpListener::bind(&self.addr).await?;
        log::info!("elkia-world listening on: {}", self.addr);

        loop {
            let (socket, addr) = listener.accept().await?;
            log::info!("Accepted connection from: {}", addr);

            let lobby_service = self.lobby_service.clone();
            let game_service = self.game_service.clone();

            tokio::spawn(async move {
                if let Err(e) = handle_connection(socket, lobby_service, game_service).await {
                    log::error!("Error handling connection from {}: {}", addr, e);
                }
            });
        }
    }
}

async fn send_character_list(framed: &mut Framed<TcpStream, GatewayCodec>, chars: &[Character]) -> Result<(), Box<dyn Error>> {
    framed.send("clist_start 0".to_string()).await?;
    for char in chars {
        let packet = format!(
            "c_info {} {} -1 {} {} {} 0 {} {} {} {} {} {} {} {} {} {} 0 0 0 0 0",
            char.name, char.id, char.class, char.level, char.hero_level,
            char.hair_color, char.hair_style, char.faction, char.reputation,
            char.dignity, char.compliment, char.job_level, char.experience,
            char.job_experience, char.hero_experience
        );
        framed.send(packet).await?;
    }
    framed.send("clist_end".to_string()).await?;
    Ok(())
}

async fn handle_connection(mut socket: TcpStream, lobby_service: Arc<dyn LobbyService>, game_service: Arc<dyn GameService>) -> Result<(), Box<dyn Error>> {
    // 1. Handshake (Session Packet)
    let (sync_cmd, user_cmd, _pass_cmd) = {
        let mut framed = Framed::new(&mut socket, SessionCodec);

        // Read SyncCommand (Packet 1)
        let pkt = framed.next().await.ok_or("Connection closed during handshake (Sync)")??;
        let sync_cmd = SyncCommand::from_str(&pkt)?;
        log::info!("Received SyncCommand: {:?}", sync_cmd);

        // Read UsernameCommand (Packet 2)
        let pkt = framed.next().await.ok_or("Connection closed during handshake (User)")??;
        let user_cmd = UsernameCommand::from_str(&pkt)?;
        log::info!("Received UsernameCommand: {:?}", user_cmd);

        // Read PasswordCommand (Packet 3)
        let pkt = framed.next().await.ok_or("Connection closed during handshake (Pass)")??;
        let pass_cmd = PasswordCommand::from_str(&pkt)?;
        log::info!("Received PasswordCommand: {:?}", pass_cmd);

        (sync_cmd, user_cmd, pass_cmd)
    };

    // 2. Game Protocol (GatewayCodec)
    let mut framed = Framed::new(socket, GatewayCodec::new(sync_cmd.code));

    // Send Character List
    let mut chars = lobby_service.get_characters(&user_cmd.username).await;

    // Create a default character if none exists (for testing)
    if chars.is_empty() {
        if let Ok(_) = lobby_service.create_character(&user_cmd.username, "Hero", 1).await {
            chars = lobby_service.get_characters(&user_cmd.username).await;
        }
    }

    send_character_list(&mut framed, &chars).await?;

    while let Some(pkt_res) = framed.next().await {
        match pkt_res {
            Ok(cmd_str) => {
                // Here we need to parse the command string into a WorldCommandPacket
                // But WorldCommandPacket::from_str expects a format like "sequence payload"
                // The gateway codec unpacks the raw bytes into a string.
                // Does the unpacked string contain the sequence number?
                // Based on WorldCommandPacket implementation:
                // fn from_str(input: &str) -> Result<Self, Self::Err> {
                //     let mut parts = input.splitn(2, ' ');
                //     let seq_str = parts.next().ok_or("Empty input")?;
                //     let sequence = seq_str.parse::<u32>().map_err(|_| "Invalid sequence")?;

                // So yes, we expect the string to start with a sequence number.

                match WorldCommandPacket::from_str(&cmd_str) {
                    Ok(cmd) => {
                        log::info!("Received Game Command: {:?}", cmd);
                        match cmd.payload {
                            WorldCommandPayload::Heartbeat => {
                                // Respond to heartbeat if needed, usually client sends periodically
                            },
                            WorldCommandPayload::Command(payload) => {
                                let parts: Vec<&str> = payload.split_whitespace().collect();
                                if let Some(tag) = parts.first() {
                                    match *tag {
                                        "select" => {
                                            // Client selects a character slot
                                            if parts.len() >= 2 {
                                                if let Ok(slot) = parts[1].parse::<usize>() {
                                                    if slot < chars.len() {
                                                        log::info!("Client selected character: {}", chars[slot].name);
                                                        framed.send("OK".to_string()).await?;
                                                    } else {
                                                        log::warn!("Client selected invalid slot: {}", slot);
                                                    }
                                                }
                                            }
                                        },
                                        "game_start" => {
                                            log::info!("Client requested game start");
                                            // Send map info, etc.
                                            // TODO: Implement map loading logic
                                        },
                                        "walk" => {
                                            if parts.len() >= 3 {
                                                if let (Ok(x), Ok(y)) = (parts[1].parse::<i32>(), parts[2].parse::<i32>()) {
                                                    game_service.walk(&user_cmd.username, x, y).await;
                                                }
                                            }
                                        },
                                        "char_new" => {
                                            // char_new name slot class ...
                                            if parts.len() >= 4 {
                                                let name = parts[1];
                                                if let Ok(class) = parts[3].parse::<i32>() {
                                                    match lobby_service.create_character(&user_cmd.username, name, class).await {
                                                        Ok(_) => {
                                                            let chars = lobby_service.get_characters(&user_cmd.username).await;
                                                            send_character_list(&mut framed, &chars).await?;
                                                        },
                                                        Err(e) => {
                                                            log::error!("Failed to create character: {}", e);
                                                        }
                                                    }
                                                }
                                            }
                                        },
                                        "say" => {
                                            if parts.len() >= 3 {
                                                let msg = parts[2..].join(" ");
                                                game_service.chat(&user_cmd.username, &msg).await;
                                            }
                                        },
                                        _ => {
                                            log::debug!("Unhandled tag: {}", tag);
                                        }
                                    }
                                }
                            }
                        }
                    },
                    Err(e) => log::error!("Failed to parse command: {}", e),
                }
            },
            Err(e) => {
                log::error!("Error reading frame: {}", e);
                break;
            }
        }
    }

    Ok(())
}
