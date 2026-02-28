use crate::net::packets::game::{GameCommandPacket, SayPacket, WalkPacket};
use crate::net::packets::lobby::{CharNewPacket, GameStartPacket, LobbyCommandPacket, SelectPacket};
use crate::net::{
    error::{Error, ParseErrorKind},
    packets::status::StatusPacket,
};
use std::fmt;
use std::str::FromStr;

#[derive(Debug, PartialEq, Clone)]
pub struct WorldCommandPacket {
    pub sequence: u32,
    pub payload: WorldCommandPayload,
}

#[derive(Debug, PartialEq, Clone)]
pub enum WorldCommandPayload {
    Heartbeat,
    Status(StatusPacket),
    Lobby(LobbyCommandPacket),
    Game(GameCommandPacket),
    Unknown(String),
}

impl FromStr for WorldCommandPacket {
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

        let payload_str = parts.next().unwrap_or("");

        // Logic to match legacy behavior: "0" or "0 ..." is Heartbeat
        let is_heartbeat = payload_str == "0" || payload_str.starts_with("0 ");

        if is_heartbeat {
            return Ok(WorldCommandPacket {
                sequence,
                payload: WorldCommandPayload::Heartbeat,
            });
        }

        let parts: Vec<&str> = payload_str.split_whitespace().collect();
        let tag = parts.first().map(|s| *s).unwrap_or("");

        let payload = match tag {
            "select" => {
                if parts.len() >= 2 {
                    if let Ok(slot) = parts[1].parse::<usize>() {
                        WorldCommandPayload::Lobby(LobbyCommandPacket::Select(SelectPacket {
                            slot,
                        }))
                    } else {
                        WorldCommandPayload::Unknown(payload_str.to_string())
                    }
                } else {
                    WorldCommandPayload::Unknown(payload_str.to_string())
                }
            }
            "game_start" => {
                WorldCommandPayload::Lobby(LobbyCommandPacket::GameStart(GameStartPacket))
            }
            "walk" => {
                if parts.len() >= 3 {
                    if let (Ok(x), Ok(y)) = (parts[1].parse::<i32>(), parts[2].parse::<i32>()) {
                        WorldCommandPayload::Game(GameCommandPacket::Walk(WalkPacket { x, y }))
                    } else {
                        WorldCommandPayload::Unknown(payload_str.to_string())
                    }
                } else {
                    WorldCommandPayload::Unknown(payload_str.to_string())
                }
            }
            "char_new" => {
                if parts.len() >= 4 {
                    let name = parts[1].to_string();
                    let slot = parts[2].parse::<usize>().unwrap_or(0);
                    if let Ok(class) = parts[3].parse::<i32>() {
                        WorldCommandPayload::Lobby(LobbyCommandPacket::CharNew(CharNewPacket {
                            name,
                            slot,
                            class,
                        }))
                    } else {
                        WorldCommandPayload::Unknown(payload_str.to_string())
                    }
                } else {
                    WorldCommandPayload::Unknown(payload_str.to_string())
                }
            }
            "say" => {
                if parts.len() >= 2 {
                    // Reconstruct message from parts[1..] to handle spaces properly?
                    // Or just take the substring after "say ".
                    // payload_str starts with "say ".
                    // Using splitn(2, ' ') on payload_str would be better.
                    let mut cmd_parts = payload_str.splitn(2, ' ');
                    cmd_parts.next(); // "say"
                    let message = cmd_parts.next().unwrap_or("").to_string();
                    WorldCommandPayload::Game(GameCommandPacket::Say(SayPacket { message }))
                } else {
                    WorldCommandPayload::Unknown(payload_str.to_string())
                }
            }
            _ => WorldCommandPayload::Unknown(payload_str.to_string()),
        };

        Ok(WorldCommandPacket { sequence, payload })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum WorldEventPacket {
    CharacterListStart(u32),
    CharacterInfo(String), // Using pre-formatted string for now or complex struct
    CharacterListEnd,
}

impl fmt::Display for WorldEventPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            WorldEventPacket::CharacterListStart(v) => write!(f, "clist_start {}", v),
            WorldEventPacket::CharacterInfo(s) => write!(f, "{}", s),
            WorldEventPacket::CharacterListEnd => write!(f, "clist_end"),
        }
    }
}
