use clap::{Parser, Subcommand};
use elkia::auth::{AuthService, SqliteAuthService};
use elkia::db;
use elkia::gateway::AuthServer;
use elkia::world::WorldServer;
use elkia::world::services::{SqliteGameService, SqliteLobbyService};
use std::error::Error;
use std::sync::Arc;
use tracing::info;
use tracing_subscriber;

/// Elkia Game Server
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand, Debug)]
enum Commands {
    /// Run database migrations
    Migrate,
    /// Run the Gateway (Auth) Server
    Gateway {
        /// Address to listen on
        #[arg(long, env = "GATEWAY_ADDR", default_value = "0.0.0.0:4000")]
        addr: String,

        /// Address of World Server to handoff to
        #[arg(long, env = "WORLD_ADDR", default_value = "127.0.0.1:5000")]
        world_addr: String,
    },
    /// Run the World Server
    World {
        /// Address to listen on
        #[arg(long, env = "WORLD_ADDR", default_value = "0.0.0.0:5000")]
        addr: String,
    },
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    tracing_subscriber::fmt::init();
    let args = Args::parse();

    let db_pool = db::connect().await?;

    match args.command {
        Commands::Migrate => {
            info!("Running database migrations...");
            db::migrate(&db_pool).await?;
            info!("Migrations completed successfully.");
        }
        Commands::Gateway { addr, world_addr } => {
            info!("Starting Elkia Gateway (Auth) Server...");
            let auth_service: Arc<dyn AuthService> = Arc::new(SqliteAuthService::new(db_pool));
            let server = Arc::new(AuthServer::new(auth_service, world_addr));
            server.run(&addr).await?;
        }
        Commands::World { addr } => {
            info!("Starting Elkia World Server...");
            // Use SqliteAuthService for session verification (read-only mostly)
            let auth_service: Arc<dyn AuthService> =
                Arc::new(SqliteAuthService::new(db_pool.clone()));

            // Sqlite services
            let lobby_service = Arc::new(SqliteLobbyService::new(db_pool.clone()));
            let game_service = Arc::new(SqliteGameService::new(db_pool));

            let server = WorldServer::new(addr, lobby_service, game_service, auth_service);
            server.run().await?;
        }
    }

    Ok(())
}
