use elkia_net::packets::session::{
    AuthInteractRequest, AuthInteractResponse, Endpoint, EndpointListEvent, LoginCommand,
};
use log::{info, warn};

pub struct GatewayServer;

impl GatewayServer {
    pub fn new() -> Self {
        Self
    }

    pub async fn handle_packet(&self, packet: AuthInteractRequest) -> Option<AuthInteractResponse> {
        match packet {
            AuthInteractRequest::Login(login_cmd) => self.handle_login(login_cmd).await,
            AuthInteractRequest::Unknown(s) => {
                warn!("Unknown packet content: {}", s);
                None
            }
        }
    }

    async fn handle_login(&self, cmd: LoginCommand) -> Option<AuthInteractResponse> {
        info!("Login attempt for user: {}", cmd.username);

        // TODO: Authenticate via gRPC (Presence Service)
        // let handoff = self.presence.auth_create_handoff_flow(...).await?;

        // TODO: Get server list via gRPC (Cluster Service)
        // let member_list = self.cluster.member_list(...).await?;

        // Mock response for now, replicating legacy behavior
        let endpoints = vec![Endpoint {
            host: "127.0.0.1".to_string(),
            port: "5000".to_string(),
            weight: 10, // calculated from population/capacity
            world_id: 1,
            channel_id: 1,
            world_name: "Elkia".to_string(),
        }];

        Some(AuthInteractResponse::EndpointList(EndpointListEvent {
            code: 0, // Success (handoff.code)
            endpoints,
        }))
    }
}
