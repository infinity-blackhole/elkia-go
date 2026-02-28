use crate::net::packets::game::GameCommandPacket;
use crate::net::packets::lobby::{LobbyCommandPacket, LobbyEventPacket};
use crate::net::{
    error::{Error, ParseErrorKind},
    packets::status::StatusEventPacket,
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
    Lobby(LobbyCommandPacket),
    Game(GameCommandPacket),
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
            "select" | "game_start" | "char_new" => {
                let lobby_cmd = payload_str.parse::<LobbyCommandPacket>()?;
                WorldCommandPayload::Lobby(lobby_cmd)
            }
            "walk" | "say" => {
                let game_cmd = payload_str.parse::<GameCommandPacket>()?;
                WorldCommandPayload::Game(game_cmd)
            }
            _ => {
                return Err(Error::parse(
                    ParseErrorKind::InvalidTag,
                    format!("Unknown command: {}", tag),
                ));
            }
        };

        Ok(WorldCommandPacket { sequence, payload })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum WorldEventPacket {
    Lobby(LobbyEventPacket),
    Status(StatusEventPacket),
}

impl fmt::Display for WorldEventPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            WorldEventPacket::Lobby(p) => write!(f, "{}", p),
            WorldEventPacket::Status(s) => write!(f, "{}", s),
        }
    }
}
