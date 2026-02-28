use crate::net::error::{Error, ParsePacketError};
use std::fmt;
use std::str::FromStr;

#[derive(Debug, PartialEq, Clone)]
pub enum LobbyCommandPacket {
    Select(SelectPacket),
    GameStart(GameStartPacket),
    CharNew(CharNewPacket),
    CharDel(CharDelPacket),
}

impl FromStr for LobbyCommandPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let tag = parts
            .next()
            .ok_or(Error::from(ParsePacketError::EmptyInput))?;
        let args = parts.next().unwrap_or("");

        match tag {
            "select" => Ok(LobbyCommandPacket::Select(args.parse()?)),
            "game_start" => Ok(LobbyCommandPacket::GameStart(args.parse()?)),
            "Char_NEW" => Ok(LobbyCommandPacket::CharNew(args.parse()?)),
            "Char_DEL" => Ok(LobbyCommandPacket::CharDel(args.parse()?)),
            _ => Err(Error::from(ParsePacketError::UnexpectedTag(
                tag.to_string(),
            ))),
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
pub struct CharDelPacket {
    pub slot: usize,
    pub password: String,
}

impl FromStr for CharDelPacket {
    type Err = Error;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.split_whitespace();
        let slot_str = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "slot".to_string(),
            )))?;
        let slot = slot_str.parse::<usize>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "slot".to_string(),
                value: slot_str.to_string(),
            })
        })?;
        let password = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "password".to_string(),
            )))?
            .to_string();
        Ok(CharDelPacket { slot, password })
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct CharNewPacket {
    pub name: String,
    pub slot: usize,
    pub gender: i32,
    pub hair_style: i32,
    pub hair_color: i32,
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
        let slot_str = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "slot".to_string(),
            )))?;
        let slot = slot_str.parse::<usize>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "slot".to_string(),
                value: slot_str.to_string(),
            })
        })?;
        let gender_str = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "gender".to_string(),
            )))?;
        let gender = gender_str.parse::<i32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "gender".to_string(),
                value: gender_str.to_string(),
            })
        })?;
        let hair_style_str = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "hair_style".to_string(),
            )))?;
        let hair_style = hair_style_str.parse::<i32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "hair_style".to_string(),
                value: hair_style_str.to_string(),
            })
        })?;
        let hair_color_str = parts
            .next()
            .ok_or(Error::from(ParsePacketError::MissingField(
                "hair_color".to_string(),
            )))?;
        let hair_color = hair_color_str.parse::<i32>().map_err(|_| {
            Error::from(ParsePacketError::InvalidField {
                field: "hair_color".to_string(),
                value: hair_color_str.to_string(),
            })
        })?;
        Ok(CharNewPacket {
            name,
            slot,
            gender,
            hair_style,
            hair_color,
        })
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
    pub id: i64,
    pub slot: i32,
    pub gender: i32,
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
            "c_info {} {} {} {} {} {} {} {} {} {} {} {} {} {} {} {} {} 0 0 0 0 0",
            self.name,
            self.id,
            self.slot,
            self.class,
            self.level,
            self.hero_level,
            self.gender,
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
