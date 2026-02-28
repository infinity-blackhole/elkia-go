use crate::net::error::{Error, ParseErrorKind};
use crate::net::packets::status::StatusPacket;
use std::fmt;
use std::str::FromStr;

#[derive(Debug, PartialEq, Clone)]
pub enum HandshakePacket {
    Sync(SyncPacket),
    Username(UsernamePacket),
    Password(PasswordPacket),
}

#[derive(Debug, PartialEq, Clone)]
pub struct SyncPacket {
    pub sequence: u32,
    pub code: u32,
}

impl FromStr for SyncPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(3, ' ');

        let seq_part = parts.next().ok_or(Error::parse(
            ParseErrorKind::Malformed,
            "Empty input".to_string(),
        ))?;

        // Remove hardcoded prefix skipping which might be incorrect for decoded Packet
        let seq_str = if seq_part.len() > 2 && seq_part.starts_with("xx") {
            &seq_part[2..]
        } else {
            seq_part
        };

        let sequence = seq_str.parse::<u32>().map_err(|_| {
            Error::parse(
                ParseErrorKind::Malformed,
                format!("Invalid sequence: {}", seq_str),
            )
        })?;

        let code_str = parts.next().ok_or(Error::parse(
            ParseErrorKind::Malformed,
            "Missing code".to_string(),
        ))?;
        let code = code_str.parse::<u32>().map_err(|_| {
            Error::parse(
                ParseErrorKind::Malformed,
                format!("Invalid code: {}", code_str),
            )
        })?;

        Ok(SyncPacket { sequence, code })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct UsernamePacket {
    pub sequence: u32,
    pub username: String,
}

impl FromStr for UsernamePacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let seq_str = parts.next().ok_or(Error::parse(
            ParseErrorKind::Malformed,
            "Empty input".to_string(),
        ))?;
        let sequence = seq_str
            .parse::<u32>()
            .map_err(|_| Error::parse(ParseErrorKind::Malformed, "Invalid sequence".to_string()))?;
        let username = parts
            .next()
            .ok_or(Error::parse(
                ParseErrorKind::Malformed,
                "Missing username".to_string(),
            ))?
            .to_string();
        Ok(UsernamePacket { sequence, username })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct PasswordPacket {
    pub sequence: u32,
    pub password: String,
}

impl FromStr for PasswordPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let seq_str = parts.next().ok_or(Error::parse(
            ParseErrorKind::Malformed,
            "Empty input".to_string(),
        ))?;
        let sequence = seq_str
            .parse::<u32>()
            .map_err(|_| Error::parse(ParseErrorKind::Malformed, "Invalid sequence".to_string()))?;
        let password = parts
            .next()
            .ok_or(Error::parse(
                ParseErrorKind::Malformed,
                "Missing password".to_string(),
            ))?
            .to_string();
        Ok(PasswordPacket { sequence, password })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum HandshakeEventPacket {
    EndpointList(EndpointListEvent),
    Status(StatusPacket),
}

impl fmt::Display for HandshakeEventPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            HandshakeEventPacket::EndpointList(ep) => write!(f, "{}", ep),
            HandshakeEventPacket::Status(s) => write!(f, "{}", s),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct Endpoint {
    pub host: String,
    pub port: String,
    pub weight: u32,
    pub world_id: u32,
    pub channel_id: u32,
    pub world_name: String,
}

impl fmt::Display for Endpoint {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{}:{}:{}:{}.{}.{}",
            self.host, self.port, self.weight, self.world_id, self.channel_id, self.world_name
        )
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct EndpointListEvent {
    pub code: u32,
    pub endpoints: Vec<Endpoint>,
}

impl fmt::Display for EndpointListEvent {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "NsTeST {} ", self.code)?;
        for ep in &self.endpoints {
            write!(f, "{} ", ep)?;
        }
        write!(f, "-1:-1:-1:10000.10000.1")
    }
}
