use crate::net::error::{Error, ParsePacketError};
use crate::net::packet::status::StatusEventPacket;
use std::fmt;
use std::str::FromStr;

#[derive(Debug, PartialEq, Clone)]
pub enum HandshakeCommandPacket {
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

        let seq_part = parts
            .next()
            .ok_or(Error::from(ParsePacketError::EmptyInput))?;

        // Remove hardcoded prefix skipping which might be incorrect for decoded Packet
        let seq_str = if seq_part.len() > 2 && seq_part.starts_with("xx") {
            &seq_part[2..]
        } else {
            seq_part
        };

        let sequence = seq_str.parse::<u32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "sequence".to_string(),
                value: seq_str.to_string(),
            })
        })?;

        let code_str = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "code".to_string(),
            )))?;
        let code = code_str.parse::<u32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "code".to_string(),
                value: code_str.to_string(),
            })
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
        let seq_str = parts.next().ok_or(Error::from(ParsePacketError::EmptyInput))?;
        let sequence = seq_str.parse::<u32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "sequence".to_string(),
                value: seq_str.to_string(),
            })
        })?;
        let username = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "username".to_string(),
            )))?
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
        let seq_str = parts.next().ok_or(Error::from(ParsePacketError::EmptyInput))?;
        let sequence = seq_str.parse::<u32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "sequence".to_string(),
                value: seq_str.to_string(),
            })
        })?;
        let password = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "password".to_string(),
            )))?
            .to_string();
        Ok(PasswordPacket { sequence, password })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum HandshakeEventPacket {
    Status(StatusEventPacket),
}

impl fmt::Display for HandshakeEventPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            HandshakeEventPacket::Status(s) => write!(f, "{}", s),
        }
    }
}
