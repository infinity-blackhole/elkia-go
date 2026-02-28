use serde::{Deserialize, Serialize};
use std::fmt;

#[derive(Debug, PartialEq, Clone)]
pub enum StatusEventPacket {
    Error(FailPacket),
    Info(InfoPacket),
}

impl fmt::Display for StatusEventPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            StatusEventPacket::Error(e) => write!(f, "{}", e),
            StatusEventPacket::Info(i) => write!(f, "{}", i),
        }
    }
}

#[derive(Debug, PartialEq, Clone)]
pub struct InfoPacket {
    pub message: String,
}

impl fmt::Display for InfoPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "info {}", self.message)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum FailCode {
    OutdatedClient = 0,
    UnexpectedError = 1,
    Maintenance = 2,
    SessionAlreadyUsed = 3,
    InvalidCredentials = 4,
    CannotAuthenticate = 5,
    UserBlocklisted = 6,
    CountryBlacklisted = 7,
    BadCase = 8,
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub struct FailPacket {
    pub code: FailCode,
}

impl FailPacket {
    pub fn new(code: FailCode) -> Self {
        Self { code }
    }
}

impl fmt::Display for FailPacket {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "failc {}", self.code as u32)
    }
}
