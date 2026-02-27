use crate::packets::status::InfoEvent;
use crate::packets::error::{Error, ErrorKind};
use std::fmt;

#[derive(Debug, PartialEq, Clone)]
pub struct LoginCommand {
    pub username: String,
    pub password: String,
    pub client_version: String,
}

#[derive(Debug, PartialEq, Clone)]
pub enum SessionCommandPacket {
    Login(LoginCommand),
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

#[derive(Debug, PartialEq, Clone)]
pub struct EndpointListEvent {
    pub code: u32,
    pub endpoints: Vec<Endpoint>,
}

#[derive(Debug, PartialEq, Clone)]
pub enum SessionEventPacket {
    Fail(Error),
    Info(InfoEvent),
    EndpointList(EndpointListEvent),
}

#[derive(Debug, PartialEq, Clone)]
pub struct UsernameCommand {
    pub sequence: u32,
    pub username: String,
}

#[derive(Debug, PartialEq, Clone)]
pub struct PasswordCommand {
    pub sequence: u32,
    pub password: String,
}

#[derive(Debug, PartialEq, Clone)]
pub struct SyncCommand {
    pub sequence: u32,
    pub code: u32,
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

impl fmt::Display for EndpointListEvent {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "NsTeST {} ", self.code)?;
        for ep in &self.endpoints {
            write!(f, "{} ", ep)?;
        }
        write!(f, "-1:-1:-1:10000.10000.1")
    }
}

impl fmt::Display for SessionEventPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            SessionEventPacket::Fail(e) => write!(f, "{}", e),
            SessionEventPacket::Info(e) => write!(f, "{}", e),
            SessionEventPacket::EndpointList(e) => write!(f, "{}", e),
        }
    }
}

use std::str::FromStr;

impl FromStr for SessionCommandPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let opcode = parts.next().ok_or(Error::new(ErrorKind::BadCase, "Empty input".to_string()))?;

        match opcode {
            "NoS0575" => {
                let rest = parts.next().ok_or(Error::new(ErrorKind::BadCase, "Missing payload".to_string()))?;
                let fields: Vec<&str> = rest.split(' ').collect();
                if fields.len() != 4 {
                    return Err(Error::new(ErrorKind::BadCase, format!("Invalid length: {}", fields.len())));
                }
                // fields[0] ignored (e.g. session id)
                let username = fields[1].to_string();
                let password = match decode_password(fields[2]) {
                    Ok(p) => p,
                    Err(e) => return Err(Error::new(ErrorKind::BadCase, format!("Password decode error: {}", e))),
                };
                let client_version = fields[3].to_string();

                Ok(SessionCommandPacket::Login(LoginCommand {
                    username,
                    password,
                    client_version,
                }))
            }
            _ => Err(Error::new(ErrorKind::BadCase, format!("Invalid opcode: {}", opcode))),
        }
    }
}

impl FromStr for UsernameCommand {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let seq_str = parts.next().ok_or(Error::new(ErrorKind::BadCase, "Empty input".to_string()))?;
        let sequence = seq_str.parse::<u32>().map_err(|_| Error::new(ErrorKind::BadCase, "Invalid sequence".to_string()))?;
        let username = parts.next().ok_or(Error::new(ErrorKind::BadCase, "Missing username".to_string()))?.to_string();
        Ok(UsernameCommand { sequence, username })
    }
}

impl FromStr for PasswordCommand {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let seq_str = parts.next().ok_or(Error::new(ErrorKind::BadCase, "Empty input".to_string()))?;
        let sequence = seq_str.parse::<u32>().map_err(|_| Error::new(ErrorKind::BadCase, "Invalid sequence".to_string()))?;
        let password = parts.next().ok_or(Error::new(ErrorKind::BadCase, "Missing password".to_string()))?.to_string();
        Ok(PasswordCommand { sequence, password })
    }
}

impl FromStr for SyncCommand {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(3, ' ');

        let seq_part = parts.next().ok_or(Error::new(ErrorKind::BadCase, "Empty input".to_string()))?;
        // Remove hardcoded prefix skipping which might be incorrect for decoded Session Packet
        let seq_str = if seq_part.len() > 2 && seq_part.starts_with("xx") {
             &seq_part[2..]
        } else {
             seq_part
        };

        let sequence = seq_str.parse::<u32>().map_err(|_| Error::new(ErrorKind::BadCase, format!("Invalid sequence: {}", seq_str)))?;

        let code_str = parts.next().ok_or(Error::new(ErrorKind::BadCase, "Missing code".to_string()))?;
        let code = code_str.parse::<u32>().map_err(|_| Error::new(ErrorKind::BadCase, format!("Invalid code: {}", code_str)))?;

        Ok(SyncCommand { sequence, code })
    }
}

impl fmt::Display for UsernameCommand {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} {}", self.sequence, self.username)
    }
}

impl fmt::Display for PasswordCommand {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{} {}", self.sequence, self.password)
    }
}

impl fmt::Display for SyncCommand {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "xx{} {}", self.sequence, self.code)
    }
}

fn decode_password(s: &str) -> Result<String, Error> {
    let bytes = s.as_bytes();
    let start = if bytes.len() % 2 == 0 { 3 } else { 4 };
    if start >= bytes.len() {
        return Ok(String::new());
    }

    let slice = &bytes[start..];

    // Take every 2nd byte
    let mut filtered = Vec::new();
    for chunk in slice.chunks(2) {
        if let Some(&b) = chunk.first() {
            filtered.push(b);
        }
    }

    String::from_utf8(filtered).map_err(|e| Error::new(ErrorKind::BadCase, e.to_string()))
}
