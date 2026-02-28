use crate::net::{error::{Error, ParseErrorKind}, packets::status::StatusEventPacket};
use std::{fmt, str::FromStr};

#[derive(Debug, PartialEq, Clone)]
pub struct LoginPacket {
    pub username: String,
    pub password: String,
    pub client_version: String,
}

#[derive(Debug, PartialEq, Clone)]
pub enum GatewayCommandPacket {
    Login(LoginPacket),
}

impl FromStr for GatewayCommandPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let tag = parts.next().ok_or(Error::parse(
            ParseErrorKind::Malformed,
            "Empty input".to_string(),
        ))?;

        match tag {
            "NoS0575" => {
                let rest = parts.next().ok_or(Error::parse(
                    ParseErrorKind::Malformed,
                    "Missing payload".to_string(),
                ))?;
                let fields: Vec<&str> = rest.split(' ').collect();
                if fields.len() != 4 {
                    return Err(Error::parse(
                        ParseErrorKind::Malformed,
                        format!("Invalid length: {}", fields.len()),
                    ));
                }
                // fields[0] ignored (e.g. session id)
                let username = fields[1].to_string();
                let password = match decode_password(fields[2]) {
                    Ok(p) => p,
                    Err(e) => {
                        return Err(Error::parse(
                            ParseErrorKind::Malformed,
                            format!("Password decode error: {}", e),
                        ));
                    }
                };
                let client_version = fields[3].to_string();

                Ok(GatewayCommandPacket::Login(LoginPacket {
                    username,
                    password,
                    client_version,
                }))
            }
            _ => Err(Error::parse(
                ParseErrorKind::InvalidTag,
                format!("Invalid tag: {}", tag),
            )),
        }
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

    String::from_utf8(filtered).map_err(|e| Error::parse(ParseErrorKind::Utf8Error, e.to_string()))
}

#[derive(Debug, PartialEq, Clone)]
pub enum GatewayEventPacket {
    EndpointList(EndpointListPacket),
    Status(StatusEventPacket),
}

impl fmt::Display for GatewayEventPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            GatewayEventPacket::EndpointList(ep) => write!(f, "{}", ep),
            GatewayEventPacket::Status(s) => write!(f, "{}", s),
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
pub struct EndpointListPacket {
    pub code: u32,
    pub endpoints: Vec<Endpoint>,
}

impl fmt::Display for EndpointListPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "NsTeST {} ", self.code)?;
        for ep in &self.endpoints {
            write!(f, "{} ", ep)?;
        }
        write!(f, "-1:-1:-1:10000.10000.1")
    }
}
