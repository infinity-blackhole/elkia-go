use crate::auth::{AuthError, AuthService};
use crate::net::codec::handshake::HandshakeCodec;
use crate::net::codec::world::WorldCodec;
use crate::net::packet::game::GameCommandPacket;
use crate::net::packet::handshake::HandshakeCommandPacket;
use crate::net::packet::lobby::{
    CharacterInfoPacket, CharacterListEndPacket, CharacterListStartPacket, LobbyCommandPacket,
    LobbyEventPacket, SelectResponsePacket,
};
use crate::net::packet::world::{WorldCommandPayload, WorldEventPacket};
use crate::world::services::{Character, GameService, LobbyService};
use futures::{SinkExt, StreamExt};
use std::fmt;
use std::sync::Arc;
use tokio::net::TcpStream;
use tokio_util::codec::Framed;
use tracing::{error, info, warn};

#[derive(Debug)]
pub enum HandshakeError {
    ConnectionClosed(String),
    ExpectedPacket(String),
    AuthError(AuthError),
    UsernameMismatch { expected: String, actual: String },
    Io(std::io::Error),
    Codec(crate::net::error::Error),
}

impl fmt::Display for HandshakeError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            HandshakeError::ConnectionClosed(s) => write!(f, "Connection closed: {}", s),
            HandshakeError::ExpectedPacket(s) => write!(f, "Expected packet: {}", s),
            HandshakeError::AuthError(e) => write!(f, "Authentication error: {}", e),
            HandshakeError::UsernameMismatch { expected, actual } => {
                write!(
                    f,
                    "Username mismatch: expected {}, got {}",
                    expected, actual
                )
            }
            HandshakeError::Io(e) => write!(f, "IO error: {}", e),
            HandshakeError::Codec(e) => write!(f, "Codec error: {}", e),
        }
    }
}

impl std::error::Error for HandshakeError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            HandshakeError::AuthError(e) => Some(e),
            HandshakeError::Io(e) => Some(e),
            HandshakeError::Codec(e) => Some(e),
            _ => None,
        }
    }
}

impl From<AuthError> for HandshakeError {
    fn from(e: AuthError) -> Self {
        HandshakeError::AuthError(e)
    }
}

impl From<std::io::Error> for HandshakeError {
    fn from(e: std::io::Error) -> Self {
        HandshakeError::Io(e)
    }
}

impl From<crate::net::error::Error> for HandshakeError {
    fn from(e: crate::net::error::Error) -> Self {
        HandshakeError::Codec(e)
    }
}

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
    ) -> Result<(TcpStream, String, u32), HandshakeError> {
        let (sync_cmd, user_cmd, pass_cmd) = {
            let mut framed = Framed::new(&mut socket, HandshakeCodec::new());

            // Read SyncCommand (Packet 1)
            let sync_cmd = match framed
                .next()
                .await
                .ok_or(HandshakeError::ConnectionClosed("Sync".into()))??
            {
                HandshakeCommandPacket::Sync(cmd) => cmd,
                _ => return Err(HandshakeError::ExpectedPacket("Sync".into())),
            };
            info!("Received SyncCommand: {:?}", sync_cmd);

            // Read UsernameCommand (Packet 2)
            let user_cmd = match framed
                .next()
                .await
                .ok_or(HandshakeError::ConnectionClosed("User".into()))??
            {
                HandshakeCommandPacket::Username(cmd) => cmd,
                _ => return Err(HandshakeError::ExpectedPacket("Username".into())),
            };
            info!("Received UsernameCommand: {:?}", user_cmd);

            // Read PasswordCommand (Packet 3)
            let pass_cmd = match framed
                .next()
                .await
                .ok_or(HandshakeError::ConnectionClosed("Pass".into()))??
            {
                HandshakeCommandPacket::Password(cmd) => cmd,
                _ => return Err(HandshakeError::ExpectedPacket("Password".into())),
            };
            info!("Received PasswordCommand: {:?}", pass_cmd);

            (sync_cmd, user_cmd, pass_cmd)
        };

        // Verify Session
        let session = self
            .auth_service
            .verify_handshake(&pass_cmd.password)
            .await?;
        if session.username != user_cmd.username {
            warn!(
                "Security Alert: Username mismatch! Packet: {}, Session: {}",
                user_cmd.username, session.username
            );
            return Err(HandshakeError::UsernameMismatch {
                expected: session.username,
                actual: user_cmd.username,
            });
        }
        info!(
            "Session verified for user: {} (ID: {})",
            session.username, session.user_id
        );

        Ok((socket, user_cmd.username, sync_cmd.code))
    }
}

#[derive(Debug)]
pub enum LobbyError {
    ConnectionClosed,
    Codec(crate::net::error::Error),
    Io(std::io::Error),
    CharacterCreation(String),
}

impl fmt::Display for LobbyError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            LobbyError::ConnectionClosed => write!(f, "Connection closed during lobby"),
            LobbyError::Codec(e) => write!(f, "Codec error: {}", e),
            LobbyError::Io(e) => write!(f, "IO error: {}", e),
            LobbyError::CharacterCreation(e) => write!(f, "Character creation failed: {}", e),
        }
    }
}

impl std::error::Error for LobbyError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            LobbyError::Codec(e) => Some(e),
            LobbyError::Io(e) => Some(e),
            _ => None,
        }
    }
}

impl From<crate::net::error::Error> for LobbyError {
    fn from(e: crate::net::error::Error) -> Self {
        LobbyError::Codec(e)
    }
}

impl From<std::io::Error> for LobbyError {
    fn from(e: std::io::Error) -> Self {
        LobbyError::Io(e)
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
    ) -> Result<(), LobbyError> {
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
            .send(WorldEventPacket::Lobby(LobbyEventPacket::CharacterListEnd(
                CharacterListEndPacket,
            )))
            .await?;
        Ok(())
    }

    pub async fn process(
        &self,
        framed: &mut Framed<TcpStream, WorldCodec>,
        username: &str,
    ) -> Result<Character, LobbyError> {
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
                    return Err(LobbyError::Codec(e));
                }
            }
        }

        Err(LobbyError::ConnectionClosed)
    }
}

#[derive(Debug)]
pub enum GameError {
    ConnectionClosed,
    Codec(crate::net::error::Error),
    Io(std::io::Error),
}

impl fmt::Display for GameError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            GameError::ConnectionClosed => write!(f, "Connection closed during game"),
            GameError::Codec(e) => write!(f, "Codec error: {}", e),
            GameError::Io(e) => write!(f, "IO error: {}", e),
        }
    }
}

impl std::error::Error for GameError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            GameError::Codec(e) => Some(e),
            GameError::Io(e) => Some(e),
            _ => None,
        }
    }
}

impl From<crate::net::error::Error> for GameError {
    fn from(e: crate::net::error::Error) -> Self {
        GameError::Codec(e)
    }
}

impl From<std::io::Error> for GameError {
    fn from(e: std::io::Error) -> Self {
        GameError::Io(e)
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
    ) -> Result<(), GameError> {
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
                    return Err(GameError::Codec(e));
                }
            }
        }

        Ok(())
    }
}
