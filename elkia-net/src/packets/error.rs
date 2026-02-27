use std::fmt;

#[derive(Debug, PartialEq, Clone)]
pub enum ErrorKind {
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

#[derive(Debug, Clone)]
pub struct Error {
    kind: ErrorKind,
    message: String,
}

impl Error {
    pub fn new(kind: ErrorKind, message: impl Into<String>) -> Self {
        Self {
            kind,
            message: message.into(),
        }
    }

    pub fn kind(&self) -> &ErrorKind {
        &self.kind
    }
}

impl From<std::io::Error> for Error {
    fn from(err: std::io::Error) -> Self {
        Self::new(ErrorKind::UnexpectedError, err.to_string())
    }
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "failc {}", self.kind.clone() as u32)
    }
}

impl std::error::Error for Error {}

impl PartialEq for Error {
    fn eq(&self, other: &Self) -> bool {
        self.kind == other.kind && self.message == other.message
    }
}
