use crate::net::error::{Error, ParseErrorKind};
use std::str::FromStr;

#[derive(Debug, PartialEq, Clone)]
pub enum GameCommandPacket {
    Walk(WalkPacket),
    Say(SayPacket),
}

impl FromStr for GameCommandPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let tag = parts.next().ok_or(Error::parse(
            ParseErrorKind::Malformed,
            "Empty input".to_string(),
        ))?;
        let args = parts.next().unwrap_or("");

        match tag {
            "walk" => Ok(GameCommandPacket::Walk(args.parse()?)),
            "say" => Ok(GameCommandPacket::Say(args.parse()?)),
            _ => Err(Error::parse(
                ParseErrorKind::InvalidTag,
                format!("Unknown game command: {}", tag),
            )),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct WalkPacket {
    pub x: i32,
    pub y: i32,
}

impl FromStr for WalkPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.split_whitespace();
        let x = parts
            .next()
            .ok_or(Error::parse(
                ParseErrorKind::Malformed,
                "Missing x coordinate".to_string(),
            ))?
            .parse::<i32>()
            .map_err(|_| {
                Error::parse(
                    ParseErrorKind::Malformed,
                    "Invalid x coordinate".to_string(),
                )
            })?;
        let y = parts
            .next()
            .ok_or(Error::parse(
                ParseErrorKind::Malformed,
                "Missing y coordinate".to_string(),
            ))?
            .parse::<i32>()
            .map_err(|_| {
                Error::parse(
                    ParseErrorKind::Malformed,
                    "Invalid y coordinate".to_string(),
                )
            })?;
        Ok(WalkPacket { x, y })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct SayPacket {
    pub message: String,
}

impl FromStr for SayPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        Ok(SayPacket {
            message: input.to_string(),
        })
    }
}
