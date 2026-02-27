use crate::codec::gateway::GatewayCodec;
use crate::codec::session::SessionCodec;
use crate::packets::command::CommandCommand;
use crate::packets::session::{PasswordCommand, SyncCommand, UsernameCommand};
use futures::stream::StreamExt;
use std::fmt;
use tokio::io::{AsyncRead, AsyncWrite};
use tokio_util::codec::Framed;

pub struct SessionClient<T> {
    framed: Framed<T, SessionCodec>,
}

impl<T: AsyncRead + AsyncWrite + Unpin> SessionClient<T> {
    pub fn new(io: T) -> Self {
        Self {
            framed: Framed::new(io, SessionCodec),
        }
    }

    pub async fn recv(&mut self) -> Result<SyncCommand, String> {
        match self.framed.next().await {
            Some(Ok(frame)) => frame.parse::<SyncCommand>(),
            Some(Err(e)) => Err(e.to_string()),
            None => Err("Connection closed".to_string()),
        }
    }

    pub fn into_inner(self) -> T {
        self.framed.into_inner()
    }
}

pub struct ChannelClient<T> {
    framed: Framed<T, GatewayCodec>,
    state: ChannelState,
}

#[derive(Debug, Clone, Copy)]
enum ChannelState {
    Username,
    Password,
    Command,
}

#[derive(Debug)]
pub enum ChannelInteractRequest {
    Username(UsernameCommand),
    Password(PasswordCommand),
    Command(CommandCommand),
}

impl fmt::Display for ChannelInteractRequest {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ChannelInteractRequest::Username(c) => write!(f, "{}", c),
            ChannelInteractRequest::Password(c) => write!(f, "{}", c),
            ChannelInteractRequest::Command(c) => write!(f, "{}", c),
        }
    }
}

impl<T: AsyncRead + AsyncWrite + Unpin> ChannelClient<T> {
    pub fn new(io: T, key: u32) -> Self {
        Self {
            framed: Framed::new(io, GatewayCodec::new(key)),
            state: ChannelState::Username,
        }
    }

    pub async fn recv(&mut self) -> Result<ChannelInteractRequest, String> {
        match self.framed.next().await {
            Some(Ok(frame)) => self.handle_frame(&frame),
            Some(Err(e)) => Err(e.to_string()),
            None => Err("Connection closed".to_string()),
        }
    }

    fn handle_frame(&mut self, frame: &str) -> Result<ChannelInteractRequest, String> {
        match self.state {
            ChannelState::Username => {
                let cmd = frame.parse::<UsernameCommand>()?;
                self.state = ChannelState::Password;
                Ok(ChannelInteractRequest::Username(cmd))
            }
            ChannelState::Password => {
                let cmd = frame.parse::<PasswordCommand>()?;
                self.state = ChannelState::Command;
                Ok(ChannelInteractRequest::Password(cmd))
            }
            ChannelState::Command => {
                let cmd = frame.parse::<CommandCommand>()?;
                Ok(ChannelInteractRequest::Command(cmd))
            }
        }
    }
}
