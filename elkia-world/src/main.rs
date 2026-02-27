mod server;
mod services;

use server::WorldServer;
use services::{InMemoryLobbyService, InMemoryGameService};
use clap::Parser;
use std::error::Error;
use std::sync::Arc;

/// Elkia World Server
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Address to listen on
    #[arg(long, env = "ADDR", default_value = "0.0.0.0:4124")]
    addr: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    env_logger::init();

    let args = Args::parse();

    log::info!("Starting elkia-world server...");

    let lobby_service = Arc::new(InMemoryLobbyService::new());
    let game_service = Arc::new(InMemoryGameService::new());

    let server = WorldServer::new(args.addr, lobby_service, game_service);
    server.run().await?;

    Ok(())
}
