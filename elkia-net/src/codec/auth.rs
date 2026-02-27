use crate::packets::session::{SessionCommandPacket, SessionEventPacket};
use crate::packets::error::{Error, ErrorKind};
use bytes::{Buf, BufMut, BytesMut};
use tokio_util::codec::{Decoder, Encoder};

pub struct AuthCodec;

impl Decoder for AuthCodec {
    type Item = SessionCommandPacket;
    type Error = Error;

    fn decode(&mut self, src: &mut BytesMut) -> Result<Option<Self::Item>, Self::Error> {
        if let Some(n) = src.iter().position(|&b| b == 0xD8) {
            let data = src.split_to(n);
            src.advance(1); // skip 0xD8

            let mut decrypted = Vec::with_capacity(data.len());
            for &b in data.iter() {
                let val = if b > 14 {
                    (b - 15) ^ 195
                } else {
                    (255 - (14 - b)) ^ 195
                };
                decrypted.push(val);
            }

            match String::from_utf8(decrypted) {
                Ok(s) => match s.parse::<SessionCommandPacket>() {
                    Ok(packet) => Ok(Some(packet)),
                    Err(e) => Err(Error::new(ErrorKind::BadCase, e.to_string())),
                },
                Err(_) => Err(Error::new(ErrorKind::BadCase, "Invalid UTF-8")),
            }
        } else {
            Ok(None)
        }
    }
}

impl Encoder<SessionEventPacket> for AuthCodec {
    type Error = Error;

    fn encode(
        &mut self,
        item: SessionEventPacket,
        dst: &mut BytesMut,
    ) -> Result<(), Self::Error> {
        let s = item.to_string();
        let bytes = s.as_bytes();

        dst.reserve(bytes.len() + 1);

        for &b in bytes {
            dst.put_u8(b.wrapping_add(15));
        }
        dst.put_u8(0x19); // Terminator

        Ok(())
    }
}
