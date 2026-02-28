use crate::auth::{AuthError, AuthService};
use crate::net::codec::gateway::GatewayCodec;
use crate::net::packet::gateway::{
    Endpoint, EndpointListPacket, GatewayCommandPacket, GatewayEventPacket, LoginPacket,
};
use crate::net::packet::status::{FailCode, FailPacket, StatusEventPacket};
use futures::{SinkExt, StreamExt};
use std::future::Future;
use std::pin::Pin;
use std::sync::Arc;
use std::task::{Context, Poll};
use tokio::net::TcpListener;
use tokio_util::codec::Framed;
use tower::Service;
use tracing::{error, info, warn};

#[derive(Clone)]
pub struct GatewayService {
    auth_service: Arc<dyn AuthService>,
    world_addr: String,
}

impl GatewayService {
    pub fn new(auth_service: Arc<dyn AuthService>, world_addr: String) -> Self {
        Self {
            auth_service,
            world_addr,
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
                let fail_code = match e {
                    AuthError::InvalidCredentials | AuthError::UserNotFound => {
                        FailCode::InvalidCredentials
                    }
                    AuthError::HandshakeExpired | AuthError::HandshakeNotFound => {
                        FailCode::CannotAuthenticate
                    }
                    AuthError::DatabaseError(_) => FailCode::UnexpectedError,
                };
                Ok(GatewayEventPacket::Status(StatusEventPacket::Error(
                    FailPacket::new(fail_code),
                )))
            }
        }
    }
}

impl Service<GatewayCommandPacket> for GatewayService {
    type Response = GatewayEventPacket;
    type Error = FailPacket;
    type Future = Pin<Box<dyn Future<Output = Result<Self::Response, Self::Error>> + Send>>;

    fn poll_ready(&mut self, _cx: &mut Context<'_>) -> Poll<Result<(), Self::Error>> {
        Poll::Ready(Ok(()))
    }

    fn call(&mut self, req: GatewayCommandPacket) -> Self::Future {
        let service = self.clone();
        Box::pin(async move {
            match req {
                GatewayCommandPacket::Login(cmd) => service.handle_login(cmd).await,
            }
        })
    }
}

pub struct GatewayServer<S> {
    service: S,
}

impl<S> GatewayServer<S>
where
    S: Service<GatewayCommandPacket, Response = GatewayEventPacket, Error = FailPacket>
        + Clone
        + Send
        + 'static,
    S::Future: Send,
{
    pub fn new(service: S) -> Self {
        Self { service }
    }

    pub async fn run(&self, addr: &str) -> Result<(), Box<dyn std::error::Error>> {
        let listener = TcpListener::bind(addr).await?;
        info!("Gateway server listening on {}", addr);

        loop {
            let (socket, _) = listener.accept().await?;
            let service = self.service.clone();
            tokio::spawn(Self::handle_connection(socket, service));
        }
    }

    async fn handle_connection(socket: tokio::net::TcpStream, mut service: S) {
        let mut framed = Framed::new(socket, GatewayCodec);

        while let Some(packet_res) = framed.next().await {
            match packet_res {
                Ok(packet) => {
                    if let Err(e) = std::future::poll_fn(|cx| service.poll_ready(cx)).await {
                        error!("Service not ready: {:?}", e);
                        break;
                    }

                    let response = match service.call(packet).await {
                        Ok(res) => res,
                        Err(e) => {
                            error!("Error handling packet: {:?}", e);
                            GatewayEventPacket::Status(StatusEventPacket::Error(FailPacket::new(
                                FailCode::UnexpectedError,
                            )))
                        }
                    };

                    if let Err(e) = framed.send(response).await {
                        error!("Failed to send response: {}", e);
                    }
                }
                Err(e) => {
                    error!("Error decoding packet: {}", e);
                    if let Err(send_err) = framed
                        .send(GatewayEventPacket::Status(StatusEventPacket::Error(
                            FailPacket::new(FailCode::BadCase),
                        )))
                        .await
                    {
                        error!("Failed to send error packet: {}", send_err);
                    }
                    break;
                }
            }
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
        ) -> Result<u32, AuthError> {
            Ok(12345)
        }

        async fn verify_handshake(&self, _handshake_id: &str) -> Result<HandshakeData, AuthError> {
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
        let mut service = GatewayService::new(auth_service, "127.0.0.1:4124".to_string());
        let cmd = LoginPacket {
            username: "test_user".to_string(),
            password: "password".to_string(),
            client_version: "1.0".to_string(),
        };

        let response = service.call(GatewayCommandPacket::Login(cmd)).await;
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
