pub mod server;

use elkia_net::codec::auth::AuthCodec;
use futures::{SinkExt, StreamExt};
use server::GatewayServer;
use std::error::Error;
use tokio::net::TcpListener;
use tokio_util::codec::Framed;

pub async fn run_server(listener: TcpListener) -> Result<(), Box<dyn Error>> {
    let addr = listener.local_addr()?;
    log::info!("elkia-gateway listening on: {}", addr);

    let server = std::sync::Arc::new(GatewayServer::new());

    loop {
        let (socket, _) = listener.accept().await?;
        let server = server.clone();

        tokio::spawn(async move {
            let mut framed = Framed::new(socket, AuthCodec);

            while let Some(result) = framed.next().await {
                match result {
                    Ok(packet) => {
                        log::info!("Received packet: {:?}", packet);

                        if let Some(response) = server.handle_packet(packet).await {
                            if let Err(e) = framed.send(response).await {
                                log::error!("Failed to send response: {}", e);
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
