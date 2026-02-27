use std::fmt;

#[derive(Debug, PartialEq, Clone)]
pub struct WorldCommandPacket {
    pub sequence: u32,
    pub payload: WorldCommandPayload,
}

#[derive(Debug, PartialEq, Clone)]
pub enum WorldCommandPayload {
    Heartbeat,
    Command(String),
}

use std::str::FromStr;
use crate::packets::error::{Error, ErrorKind};

impl FromStr for WorldCommandPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let seq_str = parts.next().ok_or(Error::new(ErrorKind::BadCase, "Empty input".to_string()))?;
        let sequence = seq_str.parse::<u32>().map_err(|_| Error::new(ErrorKind::BadCase, "Invalid sequence".to_string()))?;

        let payload_str = parts.next().unwrap_or("");

        // Logic to match legacy behavior: "0" or "0 ..." is Heartbeat
        let is_heartbeat = payload_str == "0" || payload_str.starts_with("0 ");

        if is_heartbeat {
            Ok(WorldCommandPacket {
                sequence,
                payload: WorldCommandPayload::Heartbeat,
            })
        } else {
            Ok(WorldCommandPacket {
                sequence,
                payload: WorldCommandPayload::Command(payload_str.to_string()),
            })
        }
    }
}

impl fmt::Display for WorldCommandPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.payload {
            WorldCommandPayload::Heartbeat => write!(f, "{} 0", self.sequence),
            WorldCommandPayload::Command(s) => {
                if s.is_empty() {
                    write!(f, "{}", self.sequence)
                } else {
                    write!(f, "{} {}", self.sequence, s)
                }
            }
        }
    }
}
