use crate::codec::gateway::GatewayCodec;
use crate::codec::session::SessionCodec;
use crate::packets::world::WorldCommandPacket;
use crate::packets::session::{PasswordCommand, SyncCommand, UsernameCommand};
use crate::packets::error::{Error, ErrorKind};
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

    pub async fn recv(&mut self) -> Result<SyncCommand, Error> {
        match self.framed.next().await {
            Some(Ok(frame)) => frame.parse::<SyncCommand>(),
            Some(Err(e)) => Err(Error::new(ErrorKind::BadCase, e.to_string())),
            None => Err(Error::new(ErrorKind::BadCase, "Connection closed".to_string())),
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
pub enum ChannelPayload {
    Username(UsernameCommand),
    Password(PasswordCommand),
    Command(WorldCommandPacket),
}

impl fmt::Display for ChannelPayload {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ChannelPayload::Username(c) => write!(f, "{}", c),
            ChannelPayload::Password(c) => write!(f, "{}", c),
            ChannelPayload::Command(c) => write!(f, "{}", c),
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

    pub async fn recv(&mut self) -> Result<ChannelPayload, Error> {
        match self.framed.next().await {
            Some(Ok(frame)) => self.handle_frame(&frame),
            Some(Err(e)) => Err(Error::new(ErrorKind::BadCase, e.to_string())),
            None => Err(Error::new(ErrorKind::BadCase, "Connection closed".to_string())),
        }
    }

    fn handle_frame(&mut self, frame: &str) -> Result<ChannelPayload, Error> {
        match self.state {
            ChannelState::Username => {
                let cmd = frame.parse::<UsernameCommand>()?;
                self.state = ChannelState::Password;
                Ok(ChannelPayload::Username(cmd))
            }
            ChannelState::Password => {
                let cmd = frame.parse::<PasswordCommand>()?;
                self.state = ChannelState::Command;
                Ok(ChannelPayload::Password(cmd))
            }
            ChannelState::Command => {
                let cmd = frame.parse::<WorldCommandPacket>()?;
                Ok(ChannelPayload::Command(cmd))
            }
        }
    }
}
