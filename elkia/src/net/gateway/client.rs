use crate::net::codec::handshake::HandshakeCodec;
use crate::net::codec::world::WorldCodec;
use crate::net::packets::handshake::{
    HandshakePacket, PasswordPacket, SyncPacket, UsernamePacket,
};
use crate::net::packets::status::{FailCode, FailPacket};
use crate::net::packets::world::{WorldCommandPacket, WorldCommandPayload};
use futures::stream::StreamExt;
use tokio::io::{AsyncRead, AsyncWrite};
use tokio_util::codec::Framed;

pub struct HandshakeClient<T> {
    framed: Framed<T, HandshakeCodec>,
}

impl<T: AsyncRead + AsyncWrite + Unpin> HandshakeClient<T> {
    pub fn new(io: T) -> Self {
        Self {
            framed: Framed::new(io, HandshakeCodec::new()),
        }
    }

    pub async fn recv(&mut self) -> Result<SyncPacket, FailPacket> {
        match self.framed.next().await {
            Some(Ok(packet)) => match packet {
                HandshakePacket::Sync(cmd) => Ok(cmd),
                _ => Err(FailPacket::new(FailCode::BadCase)),
            },
            Some(Err(_)) => Err(FailPacket::new(FailCode::BadCase)),
            None => Err(FailPacket::new(FailCode::BadCase)),
        }
    }

    pub fn into_inner(self) -> T {
        self.framed.into_inner()
    }
}

pub struct ChannelClient<T> {
    framed: Framed<T, WorldCodec>,
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
    Username(UsernamePacket),
    Password(PasswordPacket),
    Command(WorldCommandPacket),
}

impl<T: AsyncRead + AsyncWrite + Unpin> ChannelClient<T> {
    pub fn new(io: T, key: u32) -> Self {
        Self {
            framed: Framed::new(io, WorldCodec::new(key)),
            state: ChannelState::Username,
        }
    }

    pub async fn recv(&mut self) -> Result<ChannelPayload, FailPacket> {
        match self.framed.next().await {
            Some(Ok(frame)) => self.handle_frame(frame),
            Some(Err(_)) => Err(FailPacket::new(FailCode::BadCase)),
            None => Err(FailPacket::new(FailCode::BadCase)),
        }
    }

    fn handle_frame(&mut self, frame: WorldCommandPacket) -> Result<ChannelPayload, FailPacket> {
        match self.state {
            ChannelState::Username => {
                if let WorldCommandPayload::Unknown(s) = frame.payload {
                    let cmd = UsernamePacket {
                        sequence: frame.sequence,
                        username: s,
                    };
                    self.state = ChannelState::Password;
                    Ok(ChannelPayload::Username(cmd))
                } else {
                    Err(FailPacket::new(FailCode::BadCase))
                }
            }
            ChannelState::Password => {
                if let WorldCommandPayload::Unknown(s) = frame.payload {
                    let cmd = PasswordPacket {
                        sequence: frame.sequence,
                        password: s,
                    };
                    self.state = ChannelState::Command;
                    Ok(ChannelPayload::Password(cmd))
                } else {
                    Err(FailPacket::new(FailCode::BadCase))
                }
            }
            ChannelState::Command => Ok(ChannelPayload::Command(frame)),
        }
    }
}
