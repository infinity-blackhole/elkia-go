use bytes::{Buf, BufMut, BytesMut};
use std::io;
use tokio_util::codec::{Decoder, Encoder};

pub struct GatewayCodec {
    mode: u8,
    offset: u8,
}

impl GatewayCodec {
    pub fn new(key: u32) -> Self {
        let mode = ((key >> 6) & 0x03) as u8;
        let offset = ((key & 0xFF) + (0x40 & 0xFF)) as u8;
        Self { mode, offset }
    }

    fn decrypt(&self, data: &[u8]) -> Vec<u8> {
        let mut decrypted = Vec::with_capacity(data.len());
        for &b in data.iter() {
            let val = match self.mode {
                0 => b.wrapping_sub(self.offset),
                1 => b.wrapping_add(self.offset),
                2 => (b.wrapping_sub(self.offset)) ^ 0xC3,
                3 => (b.wrapping_add(self.offset)) ^ 0xC3,
                _ => b,
            };
            decrypted.push(val);
        }
        decrypted
    }

    fn unpack(&self, data: &[u8]) -> Result<String, io::Error> {
        let mut unpacked = Vec::new();
        let mut remaining = data;

        let permutations = b" -.0123456789n";

        while !remaining.is_empty() {
            let flag = remaining[0];
            remaining = &remaining[1..];

            if flag <= 0x7A {
                // Linear command
                let len = flag as usize;
                let actual_len = std::cmp::min(len, remaining.len());

                for i in 0..actual_len {
                    unpacked.push(remaining[i] ^ 0xFF);
                }
                remaining = &remaining[actual_len..];
            } else {
                // Compact command
                let target_len = (flag & 0x7F) as usize;
                let mut current_len = 0;

                while current_len < target_len && !remaining.is_empty() {
                    let b = remaining[0];
                    remaining = &remaining[1..];

                    let h = (b >> 4) as usize;
                    let l = (b & 0x0F) as usize;

                    if h != 0 && h != 0xF && (l == 0 || l == 0xF) {
                        if h - 1 < permutations.len() {
                            unpacked.push(permutations[h - 1]);
                            current_len += 1;
                        }
                    } else if l != 0 && l != 0xF && (h == 0 || h == 0xF) {
                        if l - 1 < permutations.len() {
                            unpacked.push(permutations[l - 1]);
                            current_len += 1;
                        }
                    } else if h != 0 && h != 0xF && l != 0 && l != 0xF {
                        if h - 1 < permutations.len() {
                            unpacked.push(permutations[h - 1]);
                            current_len += 1;
                        }
                        if current_len < target_len {
                            if l - 1 < permutations.len() {
                                unpacked.push(permutations[l - 1]);
                                current_len += 1;
                            }
                        }
                    }
                }
            }
        }

        match String::from_utf8(unpacked) {
            Ok(s) => Ok(s),
            Err(_) => Err(io::Error::new(io::ErrorKind::InvalidData, "Invalid UTF-8")),
        }
    }
}

impl Decoder for GatewayCodec {
    type Item = String;
    type Error = io::Error;

    fn decode(&mut self, src: &mut BytesMut) -> Result<Option<Self::Item>, Self::Error> {
        let delimiter = match self.mode {
            0 => 0xffu8.wrapping_add(self.offset),
            1 => 0xffu8.wrapping_sub(self.offset),
            2 => (0xffu8.wrapping_add(self.offset)) ^ 0xC3,
            3 => (0xffu8.wrapping_sub(self.offset)) ^ 0xC3,
            _ => 0xff,
        };

        if let Some(n) = src.iter().position(|&b| b == delimiter) {
            let data = src.split_to(n);
            src.advance(1); // skip delimiter
            let decrypted = self.decrypt(&data);
            let s = self.unpack(&decrypted)?;
            Ok(Some(s))
        } else {
            Ok(None)
        }
    }

    fn decode_eof(&mut self, src: &mut BytesMut) -> Result<Option<Self::Item>, Self::Error> {
        match self.decode(src)? {
            Some(frame) => Ok(Some(frame)),
            None => {
                // If we have remaining data at EOF, treat it as a frame
                if src.is_empty() {
                    Ok(None)
                } else {
                    let data = src.split_to(src.len());
                    let decrypted = self.decrypt(&data);
                    let s = self.unpack(&decrypted)?;
                    Ok(Some(s))
                }
            }
        }
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
    fn test_gateway_codec_decode_username() {
        // Input from legacy TestChannelDecodeIdentifierCommand
        let input = b"\xc6\xe4\xcb\x91\x46\xcd\xd6\xdc\xd0\xd9\xd0\xc4\x07\xd4\x49\xff\xd0\xcb\xde\xd1\xd7\xd0\xd2\xda\xc1\x70\x43\xdc\xd0\xd2\x3f\xc7\xe4\xcb\xa1\x10\x48\xd7\xd6\xdd\xc8\xd6\xc8\xd6\xf8\xc1\xa0\x41\xda\xc1\xe0\x42\xf1\xcd";

        let mut codec = GatewayCodec::new(0);
        let mut src = BytesMut::from(&input[..]);

        // First packet (Username)
        let res1 = codec.decode(&mut src).unwrap().unwrap();
        assert_eq!(res1, "60471 ricofo8350@otanhome.com");
    }
}
