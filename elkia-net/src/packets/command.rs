use std::fmt;

#[derive(Debug, PartialEq, Clone)]
pub struct CommandCommand {
    pub sequence: u32,
    pub payload: CommandPayload,
}

#[derive(Debug, PartialEq, Clone)]
pub enum CommandPayload {
    Heartbeat,
    Raw(String),
}

use std::str::FromStr;

impl FromStr for CommandCommand {
    type Err = String;

    fn from_str(input: &str) -> Result<Self, Self::Err> {
        let mut parts = input.splitn(2, ' ');
        let seq_str = parts.next().ok_or("Empty input")?;
        let sequence = seq_str.parse::<u32>().map_err(|_| "Invalid sequence")?;

        let payload_str = parts.next().unwrap_or("");

        // Logic to match legacy behavior: "0" or "0 ..." is Heartbeat
        let is_heartbeat = payload_str == "0" || payload_str.starts_with("0 ");

        if is_heartbeat {
            Ok(CommandCommand {
                sequence,
                payload: CommandPayload::Heartbeat,
            })
        } else {
            Ok(CommandCommand {
                sequence,
                payload: CommandPayload::Raw(payload_str.to_string()),
            })
        }
    }
}

impl fmt::Display for CommandCommand {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.payload {
            CommandPayload::Heartbeat => write!(f, "{} 0", self.sequence),
            CommandPayload::Raw(s) => {
                if s.is_empty() {
                    write!(f, "{}", self.sequence)
                } else {
                    write!(f, "{} {}", self.sequence, s)
                }
            }
        }
    }
}
