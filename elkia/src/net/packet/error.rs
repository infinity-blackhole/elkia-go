use std::fmt;

#[derive(Debug)]
pub enum ParsePacketError {
    EmptyInput,
    MissingField(String),
    InvalidField { field: String, value: String },
    UnexpectedTag(String),
    InvalidSequence,
    InvalidLength { expected: usize, actual: usize },
    FromUtf8(std::string::FromUtf8Error),
    ParseInt(std::num::ParseIntError),
}

impl fmt::Display for ParsePacketError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            ParsePacketError::EmptyInput => write!(f, "Empty input"),
            ParsePacketError::MissingField(s) => write!(f, "Missing field: {}", s),
            ParsePacketError::InvalidField { field, value } => {
                write!(f, "Invalid field: {}, got: {}", field, value)
            }
            ParsePacketError::UnexpectedTag(s) => write!(f, "Invalid tag: {}", s),
            ParsePacketError::InvalidSequence => write!(f, "Invalid sequence"),
            ParsePacketError::InvalidLength { expected, actual } => {
                write!(f, "Invalid length: expected {}, got {}", expected, actual)
            }
            ParsePacketError::FromUtf8(e) => write!(f, "UTF-8 error: {}", e),
            ParsePacketError::ParseInt(e) => write!(f, "Integer parse error: {}", e),
        }
    }
}

impl std::error::Error for ParsePacketError {
    fn source(&self) -> Option<&(dyn std::error::Error + 'static)> {
        match self {
            ParsePacketError::FromUtf8(e) => Some(e),
            ParsePacketError::ParseInt(e) => Some(e),
            _ => None,
        }
    }
}

impl From<std::num::ParseIntError> for ParsePacketError {
    fn from(err: std::num::ParseIntError) -> Self {
        ParsePacketError::ParseInt(err)
    }
}

impl From<std::string::FromUtf8Error> for ParsePacketError {
    fn from(err: std::string::FromUtf8Error) -> Self {
        ParsePacketError::FromUtf8(err)
    }
}
