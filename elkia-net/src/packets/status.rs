use std::fmt;

#[derive(Debug, PartialEq, Clone)]
pub struct ErrorEvent {
    pub message: String,
}

#[derive(Debug, PartialEq, Clone)]
pub struct InfoEvent {
    pub message: String,
}

impl fmt::Display for ErrorEvent {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "failc {}", self.message)
    }
}

impl fmt::Display for InfoEvent {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "info {}", self.message)
    }
}
