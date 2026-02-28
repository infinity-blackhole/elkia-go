use crate::net::packets::auth::AuthPacket;
use crate::net::packets::handshake::HandshakeEventPacket;
use bytes::{Buf, BufMut, BytesMut};
use std::io;
use tokio_util::codec::{Decoder, Encoder};

pub struct GatewayCodec;

impl Decoder for GatewayCodec {
    type Item = AuthPacket;
    type Error = io::Error;

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
                Ok(s) => match s.parse::<AuthPacket>() {
                    Ok(packet) => Ok(Some(packet)),
                    Err(e) => Err(io::Error::new(io::ErrorKind::InvalidData, e.to_string())),
                },
                Err(_) => Err(io::Error::new(io::ErrorKind::InvalidData, "Invalid UTF-8")),
            }
        } else {
            Ok(None)
        }
    }
}

impl Encoder<HandshakeEventPacket> for GatewayCodec {
    type Error = io::Error;

    fn encode(
        &mut self,
        item: HandshakeEventPacket,
        dst: &mut BytesMut,
    ) -> Result<(), Self::Error> {
        let s = match item {
            HandshakeEventPacket::EndpointList(e) => {
                let mut s = format!("NsTeST {} ", e.code);
                for (i, endpoint) in e.endpoints.iter().enumerate() {
                    if i > 0 {
                        s.push(' ');
                    }
                    s.push_str(&endpoint.to_string());
                }
                s
            }
            HandshakeEventPacket::Status(s) => format!("{}", s),
        };
        self.encode(s, dst)
    }
}

impl Encoder<String> for GatewayCodec {
    type Error = io::Error;

    fn encode(&mut self, item: String, dst: &mut BytesMut) -> Result<(), Self::Error> {
        self.encode(item.as_bytes(), dst)
    }
}

impl Encoder<&[u8]> for GatewayCodec {
    type Error = io::Error;

    fn encode(&mut self, item: &[u8], dst: &mut BytesMut) -> Result<(), Self::Error> {
        dst.reserve(item.len() + 1);

        for &b in item.iter() {
            dst.put_u8(b.wrapping_add(15));
        }
        dst.put_u8(0x19); // Terminator

        Ok(())
    }
}
