use elkia_net::packets::session::{SessionCommandPacket, SessionEventPacket, EndpointListEvent, Endpoint, LoginCommand};
use elkia_net::packets::error::{Error, ErrorKind};
use log::info;
use crate::services::AuthService;
use std::sync::Arc;

pub struct AuthServer {
    auth_service: Arc<dyn AuthService>,
}

impl AuthServer {
    pub fn new(auth_service: Arc<dyn AuthService>) -> Self {
        Self { auth_service }
    }

    pub async fn handle_packet(&self, packet: SessionCommandPacket) -> Result<SessionEventPacket, Error> {
        match packet {
            SessionCommandPacket::Login(login_cmd) => {
                self.handle_login(login_cmd).await
            }
        }
    }

    async fn handle_login(&self, cmd: LoginCommand) -> Result<SessionEventPacket, Error> {
        info!("Login attempt for user: {}", cmd.username);

        match self.auth_service.login(&cmd.username, &cmd.password).await {
            Ok(result) => {
                info!("Login successful for user: {}, session: {}", cmd.username, result.session_id);

                // Point to the elkia-world server (default port 4124)
                let endpoints = vec![
                    Endpoint {
                        host: "127.0.0.1".to_string(),
                        port: "4124".to_string(),
                        weight: 10,
                        world_id: 1,
                        channel_id: 1,
                        world_name: "Elkia".to_string(),
                    }
                ];

                Ok(SessionEventPacket::EndpointList(EndpointListEvent {
                    code: result.code,
                    endpoints,
                }))
            },
            Err(e) => {
                log::warn!("Login failed for user: {}: {}", cmd.username, e);
                Err(Error::new(ErrorKind::CannotAuthenticate, e.to_string()))
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::services::AuthResult;
    use async_trait::async_trait;

    struct MockAuthService;

    #[async_trait]
    impl AuthService for MockAuthService {
        async fn login(&self, _username: &str, _password: &str) -> Result<AuthResult, String> {
            Ok(AuthResult {
                session_id: "test-session".to_string(),
                code: 12345,
            })
        }
    }

    #[tokio::test]
    async fn test_handle_login() {
        let auth_service = Arc::new(MockAuthService);
        let server = AuthServer::new(auth_service);
        let cmd = LoginCommand {
            username: "test_user".to_string(),
            password: "password".to_string(),
            client_version: "1.0".to_string(),
        };

        let response = server.handle_packet(SessionCommandPacket::Login(cmd)).await;
        assert!(response.is_ok());

        if let SessionEventPacket::EndpointList(event) = response.unwrap() {
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
