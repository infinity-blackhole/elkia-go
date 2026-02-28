use crate::auth::AuthService;
use crate::net::codec::world::WorldCodec;
use std::error::Error;
use std::sync::Arc;
use tokio::net::{TcpListener, TcpStream};
use tokio_util::codec::Framed;
use tracing::{error, info};

pub mod services;
pub mod state;
use self::services::{GameService, LobbyService};
use self::state::{GameState, HandshakeState, LobbyState};

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

async fn handle_connection(
    socket: TcpStream,
    lobby_service: Arc<dyn LobbyService>,
    game_service: Arc<dyn GameService>,
    auth_service: Arc<dyn AuthService>,
) -> Result<(), Box<dyn Error>> {
    // 1. Handshake State
    let handshake_state = HandshakeState::new(auth_service);
    let (socket, username, code) = handshake_state.process(socket).await?;

    // 2. Lobby State
    let mut framed = Framed::new(socket, WorldCodec::new(code));
    let lobby_state = LobbyState::new(lobby_service);
    let character = lobby_state.process(&mut framed, &username).await?;

    // 3. Game State
    let game_state = GameState::new(game_service);
    game_state.process(&mut framed, &username, character).await?;

    Ok(())
}
