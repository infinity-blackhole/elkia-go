use bytes::{Buf, BufMut, BytesMut};
use std::io;
use tokio_util::codec::{Decoder, Encoder};

pub struct SessionCodec;

impl Decoder for SessionCodec {
    type Item = String;
    type Error = io::Error;

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

            match String::from_utf8(decoded) {
                Ok(s) => Ok(Some(s)),
                Err(_) => Err(io::Error::new(io::ErrorKind::InvalidData, "Invalid UTF-8")),
            }
        } else {
            Ok(None)
        }
    }
}

impl Encoder<String> for SessionCodec {
    type Error = io::Error;

    fn encode(&mut self, item: String, dst: &mut BytesMut) -> Result<(), Self::Error> {
        let len = item.len();

        dst.reserve(len + (len / 0x7E) + 2);

        for (i, &b) in item.as_bytes().iter().enumerate() {
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
    fn test_session_decode_sync_command() {
        let mut codec = SessionCodec;
        // Basic instantiation test
        let mut src = BytesMut::new();
        let _ = codec.decode(&mut src);
    }
}
