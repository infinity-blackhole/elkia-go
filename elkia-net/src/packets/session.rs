use crate::packets::status::{ErrorEvent, InfoEvent};
use std::fmt;

#[derive(Debug, PartialEq, Clone)]
pub struct LoginCommand {
    pub username: String,
    pub password: String,
    pub client_version: String,
}

#[derive(Debug, PartialEq, Clone)]
pub enum AuthInteractRequest {
    Login(LoginCommand),
    Unknown(String),
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
pub enum AuthInteractResponse {
    Error(ErrorEvent),
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

impl fmt::Display for AuthInteractResponse {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AuthInteractResponse::Error(e) => write!(f, "{}", e),
            AuthInteractResponse::Info(e) => write!(f, "{}", e),
            AuthInteractResponse::EndpointList(e) => write!(f, "{}", e),
        }
    }
}

use std::str::FromStr;

impl FromStr for AuthInteractRequest {
    type Err = String;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let opcode = parts.next().ok_or("Empty input")?;

        match opcode {
            "NoS0575" => {
                let rest = parts.next().ok_or("Missing payload")?;
                let fields: Vec<&str> = rest.split(' ').collect();
                if fields.len() != 4 {
                    return Err(format!("Invalid length: {}", fields.len()));
                }
                // fields[0] ignored (e.g. session id)
                let username = fields[1].to_string();
                let password = decode_password(fields[2])?;
                let client_version = fields[3].to_string();

                Ok(AuthInteractRequest::Login(LoginCommand {
                    username,
                    password,
                    client_version,
                }))
            }
            _ => Err(format!("Invalid opcode: {}", opcode)),
        }
    }
}

impl FromStr for UsernameCommand {
    type Err = String;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let seq_str = parts.next().ok_or("Empty input")?;
        let sequence = seq_str.parse::<u32>().map_err(|_| "Invalid sequence")?;
        let username = parts.next().ok_or("Missing username")?.to_string();
        Ok(UsernameCommand { sequence, username })
    }
}

impl FromStr for PasswordCommand {
    type Err = String;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let seq_str = parts.next().ok_or("Empty input")?;
        let sequence = seq_str.parse::<u32>().map_err(|_| "Invalid sequence")?;
        let password = parts.next().ok_or("Missing password")?.to_string();
        Ok(PasswordCommand { sequence, password })
    }
}

impl FromStr for SyncCommand {
    type Err = String;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(3, ' ');

        // fields[0][2:]
        let seq_part = parts.next().ok_or("Empty input")?;
        if seq_part.len() < 2 {
            return Err("Sequence part too short".to_string());
        }
        let seq_str = &seq_part[2..];
        let sequence = seq_str.parse::<u32>().map_err(|_| "Invalid sequence")?;

        let code_str = parts.next().ok_or("Missing code")?;
        let code = code_str.parse::<u32>().map_err(|_| "Invalid code")?;

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

fn decode_password(s: &str) -> Result<String, String> {
    let bytes = s.as_bytes();
    let start = if bytes.len() % 2 == 0 { 3 } else { 4 };
    if start >= bytes.len() {
        return Ok(String::new());
    }

    let slice = &bytes[start..];

    // Take every 2nd byte
    let mut filtered = Vec::new();
    for chunk in slice.chunks(2) {
        if !chunk.is_empty() {
            filtered.push(chunk[0]);
        }
    }

    // Hex decode
    let mut result = Vec::new();
    for chunk in filtered.chunks(2) {
        if chunk.len() == 2 {
            let s = std::str::from_utf8(chunk).map_err(|_| "Invalid UTF-8 in hex")?;
            let val = u8::from_str_radix(s, 16).map_err(|_| "Invalid hex")?;
            result.push(val);
        }
    }

    String::from_utf8(result).map_err(|_| "Invalid UTF-8 password".to_string())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_login_command() {
        // "xxx4x1" decodes to "A" (hex 41)
        // Client version needs at least one \x0b separator
        let input = "NoS0575 1234 testuser xxx4x1 1.2.3\x0b0";

        let result = input.parse::<AuthInteractRequest>().unwrap();

        if let AuthInteractRequest::Login(cmd) = result {
            assert_eq!(cmd.username, "testuser");
            assert_eq!(cmd.password, "A");
            assert_eq!(cmd.client_version, "1.2.3\x0b0");
        } else {
            panic!("Expected Login command");
        }
    }

    #[test]
    fn test_parse_username_command() {
        let input = "12345 testuser";
        let cmd = input.parse::<UsernameCommand>().unwrap();
        assert_eq!(cmd.sequence, 12345);
        assert_eq!(cmd.username, "testuser");
    }
}
