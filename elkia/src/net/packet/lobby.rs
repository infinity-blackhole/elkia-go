use crate::net::error::{Error, ParsePacketError};
use std::fmt;
use std::str::FromStr;

#[derive(Debug, PartialEq, Clone)]
pub enum LobbyCommandPacket {
    Select(SelectPacket),
    GameStart(GameStartPacket),
    CharNew(CharNewPacket),
}

impl FromStr for LobbyCommandPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let tag = parts.next().ok_or(Error::from(ParsePacketError::EmptyInput))?;
        let args = parts.next().unwrap_or("");

        match tag {
            "select" => Ok(LobbyCommandPacket::Select(args.parse()?)),
            "game_start" => Ok(LobbyCommandPacket::GameStart(args.parse()?)),
            "char_new" => Ok(LobbyCommandPacket::CharNew(args.parse()?)),
            _ => Err(Error::from(ParsePacketError::UnexpectedTag(tag.to_string()))),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct SelectPacket {
    pub slot: usize,
}

impl FromStr for SelectPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let slot_str = input.trim();
        let slot = slot_str.parse::<usize>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "slot".to_string(),
                value: slot_str.to_string(),
            })
        })?;
        Ok(SelectPacket { slot })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct GameStartPacket;

impl FromStr for GameStartPacket {
    type Err = Error;

    fn from_str(_input: &str) -> Result<Self, Self::Err> {
        Ok(GameStartPacket)
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct CharNewPacket {
    pub name: String,
    pub slot: usize,
    pub class: i32,
}

impl FromStr for CharNewPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.split_whitespace();
        let name = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "name".to_string(),
            )))?
            .to_string();
        let slot_str = parts.next().ok_or(Error::from(ParsePacketError::MissingField(
            "slot".to_string(),
        )))?;
        let slot = slot_str.parse::<usize>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "slot".to_string(),
                value: slot_str.to_string(),
            })
        })?;
        let class_str = parts.next().ok_or(Error::from(ParsePacketError::MissingField(
            "class".to_string(),
        )))?;
        let class = class_str.parse::<i32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "class".to_string(),
                value: class_str.to_string(),
            })
        })?;
        Ok(CharNewPacket { name, slot, class })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub enum LobbyEventPacket {
    CharacterListStart(CharacterListStartPacket),
    CharacterInfo(CharacterInfoPacket),
    CharacterListEnd(CharacterListEndPacket),
    SelectResponse(SelectResponsePacket),
}

impl fmt::Display for LobbyEventPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            LobbyEventPacket::CharacterListStart(p) => write!(f, "{}", p),
            LobbyEventPacket::CharacterInfo(p) => write!(f, "{}", p),
            LobbyEventPacket::CharacterListEnd(p) => write!(f, "{}", p),
            LobbyEventPacket::SelectResponse(p) => write!(f, "{}", p),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct CharacterListStartPacket {
    pub sequence: u32,
}

impl fmt::Display for CharacterListStartPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "clist_start {}", self.sequence)
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct CharacterInfoPacket {
    pub name: String,
    pub id: String,
    pub class: i32,
    pub level: i32,
    pub hero_level: i32,
    pub hair_color: i32,
    pub hair_style: i32,
    pub faction: i32,
    pub reputation: i32,
    pub dignity: i32,
    pub compliment: i32,
    pub job_level: i32,
    pub experience: i32,
    pub job_experience: i32,
    pub hero_experience: i32,
}

impl fmt::Display for CharacterInfoPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "c_info {} {} -1 {} {} {} 0 {} {} {} {} {} {} {} {} {} {} 0 0 0 0 0",
            self.name,
            self.id,
            self.class,
            self.level,
            self.hero_level,
            self.hair_color,
            self.hair_style,
            self.faction,
            self.reputation,
            self.dignity,
            self.compliment,
            self.job_level,
            self.experience,
            self.job_experience,
            self.hero_experience
        )
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct CharacterListEndPacket;

impl fmt::Display for CharacterListEndPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "clist_end")
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct SelectResponsePacket;

impl fmt::Display for SelectResponsePacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "OK")
    }
}
