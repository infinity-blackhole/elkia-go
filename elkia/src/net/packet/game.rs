use crate::net::error::{Error, ParsePacketError};
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
        let tag = parts
            .next()
            .ok_or(Error::from(ParsePacketError::EmptyInput))?;
        let args = parts.next().unwrap_or("");

        match tag {
            "walk" => Ok(GameCommandPacket::Walk(args.parse()?)),
            "say" => Ok(GameCommandPacket::Say(args.parse()?)),
            _ => Err(Error::from(ParsePacketError::UnexpectedTag(tag.to_string()))),
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
        let x_str = parts.next().ok_or(Error::from(ParsePacketError::MissingField(
            "x coordinate".to_string(),
        )))?;
        let x = x_str.parse::<i32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "x coordinate".to_string(),
                value: x_str.to_string(),
            })
        })?;
        let y_str = parts.next().ok_or(Error::from(ParsePacketError::MissingField(
            "y coordinate".to_string(),
        )))?;
        let y = y_str.parse::<i32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "y coordinate".to_string(),
                value: y_str.to_string(),
            })
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
