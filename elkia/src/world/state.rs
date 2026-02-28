use crate::auth::{AuthError, AuthService};
use crate::game::GameService;
use crate::lobby::LobbyService;
use crate::net::codec::handshake::HandshakeCodec;
use crate::net::codec::world::WorldCodec;
use crate::net::packet::game::GameCommandPacket;
use crate::net::packet::handshake::{HandshakeCommandPacket, HandshakeEventPacket};
use crate::net::packet::lobby::{
    CharacterInfoPacket, LobbyCommandPacket, LobbyEventPacket, SelectResponsePacket,
};
use crate::net::packet::status::{FailCode, FailPacket, StatusEventPacket};
use crate::net::packet::world::{WorldCommandPayload, WorldEventPacket};
use crate::net::utils::lobby::send_character_list;
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
    ) -> Result<(TcpStream, i64, u32), HandshakeError> {
        let (sync, user, pass) = {
            let mut framed = Framed::new(&mut socket, HandshakeCodec::new());

            // Read SyncCommand (Packet 1)
            let sync = match framed
                .next()
                .await
                .ok_or(HandshakeError::ConnectionClosed("Sync".into()))??
            {
                HandshakeCommandPacket::Sync(cmd) => cmd,
                _ => return Err(HandshakeError::ExpectedPacket("Sync".into())),
            };
            info!("Received SyncCommand: {:?}", sync);

            // Read UsernameCommand (Packet 2)
            let user = match framed
                .next()
                .await
                .ok_or(HandshakeError::ConnectionClosed("User".into()))??
            {
                HandshakeCommandPacket::Username(cmd) => cmd,
                _ => return Err(HandshakeError::ExpectedPacket("Username".into())),
            };
            info!("Received UsernameCommand: {:?}", user);

            // Read PasswordCommand (Packet 3)
            let pass = match framed
                .next()
                .await
                .ok_or(HandshakeError::ConnectionClosed("Pass".into()))??
            {
                HandshakeCommandPacket::Password(cmd) => cmd,
                _ => return Err(HandshakeError::ExpectedPacket("Password".into())),
            };
            info!("Received PasswordCommand: {:?}", pass);

            (sync, user, pass)
        };

        // Verify World Login (Credentials + Session)
        let session_id = match self
            .auth_service
            .activate_session(&user.username, &pass.password, sync.code)
            .await
        {
            Ok(id) => id,
            Err(e) => {
                warn!("World login failed for user: {}: {:?}", user.username, e);

                let fail_code = match e {
                    AuthError::InvalidCredentials | AuthError::UserNotFound => {
                        FailCode::InvalidCredentials
                    }
                    AuthError::ActiveSession => FailCode::SessionAlreadyUsed,
                    AuthError::HandshakeExpired | AuthError::HandshakeNotFound => {
                        FailCode::CannotAuthenticate
                    }
                    _ => FailCode::UnexpectedError,
                };

                let mut framed = Framed::new(&mut socket, HandshakeCodec::new());
                let packet = HandshakeEventPacket::Status(StatusEventPacket::Error(
                    FailPacket::new(fail_code),
                ));
                let _ = framed.send(packet).await;
                return Err(HandshakeError::AuthError(e));
            }
        };
        Ok((socket, session_id, sync.code))
    }
}

#[derive(Debug)]
pub enum LobbyStateError {
    ConnectionClosed,
    Net(crate::net::error::Error),
    Io(std::io::Error),
    Service(crate::lobby::LobbyError),
}

impl fmt::Display for LobbyStateError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            LobbyStateError::ConnectionClosed => write!(f, "Connection closed during lobby"),
            LobbyStateError::Net(e) => write!(f, "Codec error: {}", e),
            LobbyStateError::Io(e) => write!(f, "IO error: {}", e),
            LobbyStateError::Service(e) => write!(f, "Lobby service error: {}", e),
        }
    }
}

impl std::error::Error for LobbyStateError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            LobbyStateError::Net(e) => Some(e),
            LobbyStateError::Io(e) => Some(e),
            LobbyStateError::Service(e) => Some(e),
            _ => None,
        }
    }
}

impl From<crate::net::error::Error> for LobbyStateError {
    fn from(e: crate::net::error::Error) -> Self {
        LobbyStateError::Net(e)
    }
}

impl From<std::io::Error> for LobbyStateError {
    fn from(e: std::io::Error) -> Self {
        LobbyStateError::Io(e)
    }
}

impl From<crate::lobby::LobbyError> for LobbyStateError {
    fn from(e: crate::lobby::LobbyError) -> Self {
        LobbyStateError::Service(e)
    }
}

pub struct LobbyState {
    lobby_service: Arc<dyn LobbyService>,
    auth_service: Arc<dyn AuthService>,
}

impl LobbyState {
    pub fn new(lobby_service: Arc<dyn LobbyService>, auth_service: Arc<dyn AuthService>) -> Self {
        Self {
            lobby_service,
            auth_service,
        }
    }

    pub async fn process(
        &self,
        framed: &mut Framed<TcpStream, WorldCodec>,
        session_id: i64,
    ) -> Result<(), LobbyStateError> {
        let chars = self
            .lobby_service
            .list_characters(session_id)
            .filter_map(|res| async move {
                match res {
                    Ok(c) => Some(CharacterInfoPacket::from(c)),
                    Err(e) => {
                        warn!("Error listing characters: {}", e);
                        None
                    }
                }
            })
            .boxed();
        send_character_list(framed, chars).await?;

        while let Some(pkt_res) = framed.next().await {
            match pkt_res {
                Ok(cmd) => {
                    info!("Lobby Received Command: {:?}", cmd);
                    match cmd.payload {
                        WorldCommandPayload::Heartbeat => {
                            if let Err(e) = self.auth_service.refresh_session(session_id).await {
                                warn!("Failed to update heartbeat: {}", e);
                            }
                        }
                        WorldCommandPayload::Lobby(lobby) => match lobby {
                            LobbyCommandPacket::Select(pkt) => {
                                info!("Client selecting slot: {}", pkt.slot);
                                match self
                                    .lobby_service
                                    .select_character(session_id, pkt.slot)
                                    .await
                                {
                                    Ok(char_id) => {
                                        info!("Client selected character ID: {}", char_id);
                                        framed
                                            .send(WorldEventPacket::Lobby(
                                                LobbyEventPacket::SelectResponse(
                                                    SelectResponsePacket,
                                                ),
                                            ))
                                            .await?;
                                    }
                                    Err(e) => {
                                        warn!("Failed to select character (invalid slot?): {}", e);
                                    }
                                }
                            }
                            LobbyCommandPacket::GameStart(_) => {
                                return Ok(());
                            }
                            LobbyCommandPacket::CharNew(pkt) => {
                                let chars = self
                                    .lobby_service
                                    .create_character(
                                        session_id,
                                        &pkt.name,
                                        pkt.slot as i32,
                                        pkt.gender,
                                        pkt.hair_style,
                                        pkt.hair_color,
                                    )
                                    .filter_map(|res| async move {
                                        match res {
                                            Ok(c) => Some(CharacterInfoPacket::from(c)),
                                            Err(e) => {
                                                warn!("Error creating character: {}", e);
                                                None
                                            }
                                        }
                                    })
                                    .boxed();
                                send_character_list(framed, chars).await?;
                            }
                            LobbyCommandPacket::CharDel(pkt) => {
                                let chars = self
                                    .lobby_service
                                    .delete_character(session_id, pkt.slot, &pkt.password)
                                    .filter_map(|res| async move {
                                        match res {
                                            Ok(c) => Some(CharacterInfoPacket::from(c)),
                                            Err(e) => {
                                                warn!("Error deleting character: {}", e);
                                                None
                                            }
                                        }
                                    })
                                    .boxed();
                                send_character_list(framed, chars).await?;
                            }
                        },
                        WorldCommandPayload::Game(_) => {
                            warn!("Received game command in lobby state, ignoring");
                        }
                    }
                }
                Err(e) => return Err(LobbyStateError::Net(e)),
            }
        }
        Err(LobbyStateError::ConnectionClosed)
    }
}

#[derive(Debug)]
pub enum GameStateError {
    ConnectionClosed,
    Codec(crate::net::error::Error),
    Io(std::io::Error),
    Service(crate::game::GameError),
}

impl fmt::Display for GameStateError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            GameStateError::ConnectionClosed => write!(f, "Connection closed during game"),
            GameStateError::Codec(e) => write!(f, "Codec error: {}", e),
            GameStateError::Io(e) => write!(f, "IO error: {}", e),
            GameStateError::Service(e) => write!(f, "Game service error: {}", e),
        }
    }
}

impl std::error::Error for GameStateError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            GameStateError::Codec(e) => Some(e),
            GameStateError::Io(e) => Some(e),
            GameStateError::Service(e) => Some(e),
            _ => None,
        }
    }
}

impl From<crate::net::error::Error> for GameStateError {
    fn from(e: crate::net::error::Error) -> Self {
        GameStateError::Codec(e)
    }
}

impl From<std::io::Error> for GameStateError {
    fn from(e: std::io::Error) -> Self {
        GameStateError::Io(e)
    }
}

impl From<crate::game::GameError> for GameStateError {
    fn from(e: crate::game::GameError) -> Self {
        GameStateError::Service(e)
    }
}

pub struct GameState {
    game_service: Arc<dyn GameService>,
    auth_service: Arc<dyn AuthService>,
}

impl GameState {
    pub fn new(game_service: Arc<dyn GameService>, auth_service: Arc<dyn AuthService>) -> Self {
        Self {
            game_service,
            auth_service,
        }
    }

    pub async fn process(
        &self,
        framed: &mut Framed<TcpStream, WorldCodec>,
        session_id: i64,
    ) -> Result<(), GameStateError> {
        while let Some(pkt_res) = framed.next().await {
            match pkt_res {
                Ok(cmd) => {
                    info!("Game Received Command: {:?}", cmd);
                    match cmd.payload {
                        WorldCommandPayload::Heartbeat => {
                            if let Err(e) = self.auth_service.refresh_session(session_id).await {
                                warn!("Failed to update heartbeat: {}", e);
                            }
                        }
                        WorldCommandPayload::Lobby(_) => {
                            warn!("Received lobby command in game state, ignoring");
                        }
                        WorldCommandPayload::Game(game) => match game {
                            GameCommandPacket::Walk(pkt) => {
                                if let Err(e) =
                                    self.game_service.walk(session_id, pkt.x, pkt.y).await
                                {
                                    error!("Failed to process walk command: {}", e);
                                    // Depending on severity, we might want to return Err or just log
                                }
                            }
                            GameCommandPacket::Say(pkt) => {
                                if let Err(e) =
                                    self.game_service.chat(session_id, &pkt.message).await
                                {
                                    error!("Failed to process say command: {}", e);
                                }
                            }
                        },
                    }
                }
                Err(e) => {
                    return Err(GameStateError::Codec(e));
                }
            }
        }

        Ok(())
    }
}
