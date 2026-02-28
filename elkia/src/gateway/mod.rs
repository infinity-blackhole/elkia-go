use crate::auth::AuthService;
use crate::net::codec::gateway::GatewayCodec;
use crate::net::packets::gateway::{GatewayCommandPacket, LoginPacket,Endpoint, EndpointListPacket, GatewayEventPacket};
use crate::net::packets::status::{FailCode, FailPacket, StatusEventPacket};
use futures::{SinkExt, StreamExt};
use std::sync::Arc;
use tokio::net::TcpListener;
use tokio_util::codec::Framed;
use tracing::{error, info, warn};

pub struct AuthServer {
    auth_service: Arc<dyn AuthService>,
    world_addr: String,
}

impl AuthServer {
    pub fn new(auth_service: Arc<dyn AuthService>, world_addr: String) -> Self {
        Self {
            auth_service,
            world_addr,
        }
    }

    pub async fn run(&self, addr: &str) -> Result<(), Box<dyn std::error::Error>> {
        let listener = TcpListener::bind(addr).await?;
        info!("Auth server listening on {}", addr);

        loop {
            let (socket, _) = listener.accept().await?;
            let server = self.clone();
            tokio::spawn(async move {
                if let Err(e) = server.handle_connection(socket).await {
                    error!("Connection error: {}", e);
                }
            });
        }
    }

    async fn handle_connection(
        &self,
        mut socket: tokio::net::TcpStream,
    ) -> Result<(), Box<dyn std::error::Error>> {
        let mut framed = Framed::new(&mut socket, GatewayCodec);

        while let Some(packet) = framed.next().await {
            match packet {
                Ok(packet) => match self.handle_packet(packet).await {
                    Ok(response) => {
                        framed.send(response).await?;
                    }
                    Err(e) => {
                        error!("Error handling packet: {}", e);
                        if let Err(send_err) = framed
                            .send(GatewayEventPacket::Status(StatusEventPacket::Error(e)))
                            .await
                        {
                            error!("Failed to send error packet: {}", send_err);
                        }
                        // Usually, auth failure implies disconnection or retry.
                        // For now, we don't break, allowing retry if client supports it.
                        // break;
                    }
                },
                Err(e) => {
                    error!("Error decoding packet: {}", e);
                    break;
                }
            }
        }
        Ok(())
    }

    pub async fn handle_packet(
        &self,
        packet: GatewayCommandPacket,
    ) -> Result<GatewayEventPacket, FailPacket> {
        match packet {
            GatewayCommandPacket::Login(login_cmd) => self.handle_login(login_cmd).await,
        }
    }

    async fn handle_login(&self, cmd: LoginPacket) -> Result<GatewayEventPacket, FailPacket> {
        info!("Login attempt for user: {}", cmd.username);

        match self
            .auth_service
            .create_handshake_flow(&cmd.username, &cmd.password)
            .await
        {
            Ok(code) => {
                info!(
                    "Login successful for user: {}, code: {}",
                    cmd.username, code
                );

                // Parse host and port from world_addr
                let parts: Vec<&str> = self.world_addr.split(':').collect();
                let host = parts.get(0).unwrap_or(&"127.0.0.1").to_string();
                let port = parts.get(1).unwrap_or(&"5000").to_string();

                // Point to the elkia-world server
                let endpoints = vec![Endpoint {
                    host,
                    port,
                    weight: 10,
                    world_id: 1,
                    channel_id: 1,
                    world_name: "Elkia".to_string(),
                }];

                Ok(GatewayEventPacket::EndpointList(EndpointListPacket {
                    code,
                    endpoints,
                }))
            }
            Err(e) => {
                warn!("Login failed for user: {}: {}", cmd.username, e);
                Err(FailPacket::new(FailCode::CannotAuthenticate))
            }
        }
    }
}

impl Clone for AuthServer {
    fn clone(&self) -> Self {
        Self {
            auth_service: self.auth_service.clone(),
            world_addr: self.world_addr.clone(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::auth::{AuthService, HandshakeData};
    use async_trait::async_trait;

    struct MockAuthService;

    #[async_trait]
    impl AuthService for MockAuthService {
        async fn create_handshake_flow(
            &self,
            _username: &str,
            _password: &str,
        ) -> Result<u32, String> {
            Ok(12345)
        }

        async fn verify_handshake(&self, _handshake_id: &str) -> Result<HandshakeData, String> {
            Ok(HandshakeData {
                id: "test-handshake".to_string(),
                user_id: "test-user".to_string(),
                username: "test_user".to_string(),
            })
        }
    }

    #[tokio::test]
    async fn test_handle_login() {
        let auth_service = Arc::new(MockAuthService);
        let server = AuthServer::new(auth_service, "127.0.0.1:4124".to_string());
        let cmd = LoginPacket {
            username: "test_user".to_string(),
            password: "password".to_string(),
            client_version: "1.0".to_string(),
        };

        let response = server.handle_packet(GatewayCommandPacket::Login(cmd)).await;
        assert!(response.is_ok());

        if let GatewayEventPacket::EndpointList(event) = response.unwrap() {
            assert_eq!(event.code, 12345);
            assert!(!event.endpoints.is_empty());
            let endpoint = &event.endpoints[0];
            assert_eq!(endpoint.host, "127.0.0.1");
            assert_eq!(endpoint.port, "4124");
        } else {
            panic!("Expected EndpointList response");
        }
    }
}
