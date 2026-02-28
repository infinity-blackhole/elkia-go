use crate::net::error::{Error, ParseErrorKind};
use std::str::FromStr;

#[derive(Debug, PartialEq, Clone)]
pub struct LoginPacket {
    pub username: String,
    pub password: String,
    pub client_version: String,
}

#[derive(Debug, PartialEq, Clone)]
pub enum AuthPacket {
    Login(LoginPacket),
}

impl FromStr for AuthPacket {
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

                Ok(AuthPacket::Login(LoginPacket {
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
