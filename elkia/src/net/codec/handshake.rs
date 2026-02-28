use crate::net::error::{Error, ParsePacketError};
use crate::net::packet::handshake::{
    HandshakeCommandPacket, HandshakeEventPacket, PasswordPacket, SyncPacket, UsernamePacket,
};
use bytes::{Buf, BufMut, BytesMut};
use std::io;
use tokio_util::codec::{Decoder, Encoder};

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum State {
    Sync,
    Username,
    Password,
    Done,
}

pub struct HandshakeCodec {
    state: State,
}

impl HandshakeCodec {
    pub fn new() -> Self {
        Self { state: State::Sync }
    }
}

impl Default for HandshakeCodec {
    fn default() -> Self {
        Self::new()
    }
}

impl Decoder for HandshakeCodec {
    type Item = HandshakeCommandPacket;
    type Error = Error;

    fn decode(&mut self, src: &mut BytesMut) -> Result<Option<Self::Item>, Self::Error> {
        if let Some(n) = src.iter().position(|&b| b == 0x0E) {
            // 0x0E delimiter
            let data = src.split_to(n);
            src.advance(1); // skip delimiter

            let mut decoded = Vec::with_capacity(data.len() * 2);

            for &b in data.iter() {
                let first_byte = b.wrapping_sub(0xF);
                let second_byte = first_byte & 0xF0;
                let second_key = second_byte >> 4;
                let first_key = first_byte.wrapping_sub(second_byte);

                for key in [second_key, first_key] {
                    let char_byte = match key {
                        0 | 1 => b' ',
                        2 => b'-',
                        3 => b'.',
                        _ => 0x2C + key,
                    };
                    decoded.push(char_byte);
                }
            }

            let s = match String::from_utf8(decoded) {
                Ok(s) => s,
                Err(e) => return Err(Error::from(ParsePacketError::FromUtf8(e))),
            };

            let packet = match self.state {
                State::Sync => {
                    use std::str::FromStr;
                    let cmd = SyncPacket::from_str(&s)?;
                    self.state = State::Username;
                    HandshakeCommandPacket::Sync(cmd)
                }
                State::Username => {
                    use std::str::FromStr;
                    let cmd = UsernamePacket::from_str(&s)?;
                    self.state = State::Password;
                    HandshakeCommandPacket::Username(cmd)
                }
                State::Password => {
                    use std::str::FromStr;
                    let cmd = PasswordPacket::from_str(&s)?;
                    self.state = State::Done;
                    HandshakeCommandPacket::Password(cmd)
                }
                State::Done => {
                    // Should not happen in normal flow, but maybe reconnect/error
                    return Err(Error::from(ParsePacketError::InvalidSequence));
                }
            };

            Ok(Some(packet))
        } else {
            Ok(None)
        }
    }
}

impl Encoder<HandshakeEventPacket> for HandshakeCodec {
    type Error = io::Error;

    fn encode(
        &mut self,
        item: HandshakeEventPacket,
        dst: &mut BytesMut,
    ) -> Result<(), Self::Error> {
        self.encode(item.to_string(), dst)
    }
}

impl Encoder<String> for HandshakeCodec {
    type Error = io::Error;

    fn encode(&mut self, item: String, dst: &mut BytesMut) -> Result<(), Self::Error> {
        self.encode(item.as_bytes(), dst)
    }
}

impl Encoder<&[u8]> for HandshakeCodec {
    type Error = io::Error;

    fn encode(&mut self, item: &[u8], dst: &mut BytesMut) -> Result<(), Self::Error> {
        let len = item.len();

        dst.reserve(len + (len / 0x7E) + 2);

        for (i, &b) in item.iter().enumerate() {
            if i % 0x7E == 0 {
                let remaining = len - i;
                let chunk_len = std::cmp::min(remaining, 0x7E) as u8;
                dst.put_u8(chunk_len);
            }
            dst.put_u8(!b);
        }
        dst.put_u8(0x19);

        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_handshake_decode_sync_command() {
        let mut codec = HandshakeCodec::new();
        // Basic instantiation test
        let mut src = BytesMut::new();
        let _ = codec.decode(&mut src);
    }
}
