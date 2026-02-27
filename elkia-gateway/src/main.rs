mod server;
mod services;

use std::error::Error;
use std::sync::Arc;
use tokio::net::TcpListener;
use tokio_util::codec::Framed;
use futures::{SinkExt, StreamExt};
use elkia_net::codec::auth::AuthCodec;
use elkia_net::packets::session::SessionEventPacket;
use server::AuthServer;
use services::InMemoryAuthService;
use clap::Parser;

/// Elkia Gateway Server
#[derive(Parser, Debug)]
#[command(author, version, about, long_about = None)]
struct Args {
    /// Address to listen on
    #[arg(long, env = "ADDR", default_value = "0.0.0.0:4000")]
    addr: String,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn Error>> {
    env_logger::init();

    let args = Args::parse();

    let listener = TcpListener::bind(&args.addr).await?;
    log::info!("elkia-gateway listening on: {}", args.addr);

    // In a real implementation, we would share the server instance (e.g. for DB pool)
    // or instantiate per connection if lightweight. For now, one global instance.
    let auth_service = Arc::new(InMemoryAuthService::new());
    let server = Arc::new(AuthServer::new(auth_service));

    loop {
        let (socket, _) = listener.accept().await?;
        let server = server.clone();

        tokio::spawn(async move {
            let mut framed = Framed::new(socket, AuthCodec);

            while let Some(result) = framed.next().await {
                match result {
                    Ok(packet) => {
                        log::info!("Received packet: {:?}", packet);

                        match server.handle_packet(packet).await {
                            Ok(response) => {
                                if let Err(e) = framed.send(response).await {
                                    log::error!("Failed to send response: {}", e);
                                    break;
                                }
                            }
                            Err(e) => {
                                if let Err(e) = framed.send(SessionEventPacket::Fail(e)).await {
                                    log::error!("Failed to send error response: {}", e);
                                    break;
                                }
                                break;
                            }
                        }
                    }
                    Err(e) => {
                        log::error!("Error decoding packet: {}", e);
                        break;
                    }
                }
            }
            log::info!("Connection closed");
        });
    }
}
