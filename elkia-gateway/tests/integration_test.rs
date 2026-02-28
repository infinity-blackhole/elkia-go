use bytes::{Buf, BufMut, BytesMut};
use elkia_net::packets::session::{AuthInteractRequest, LoginCommand};
use futures::{SinkExt, StreamExt};
use std::error::Error;
use std::io;
use tokio::net::{TcpListener, TcpStream};
use tokio_util::codec::{Decoder, Encoder, Framed};

// Define a client-side codec for testing
pub struct TestClientCodec;

impl Encoder<AuthInteractRequest> for TestClientCodec {
    type Error = io::Error;

    fn encode(&mut self, item: AuthInteractRequest, dst: &mut BytesMut) -> Result<(), Self::Error> {
        // Convert request to string (we need Display impl or manual formatting)
        let s = match item {
            // NoS0575 <session_id> <username> <password> <version>
            // We use "0" as dummy session_id as fields[0] is ignored by server
            // Password must be encoded: hex string interleaved with 'x', prefixed with "xxx"
            AuthInteractRequest::Login(cmd) => {
                let hex_pass: String = cmd.password.bytes().map(|b| format!("{:02X}", b)).collect();
                let mut encoded_pass = String::from("xxx");
                for (i, c) in hex_pass.chars().enumerate() {
                    if i > 0 {
                        encoded_pass.push('x');
                    }
                    encoded_pass.push(c);
                }
                format!(
                    "NoS0575 0 {} {} {}",
                    cmd.username, encoded_pass, cmd.client_version
                )
            }
            _ => {
                return Err(io::Error::new(
                    io::ErrorKind::InvalidInput,
                    "Unsupported packet for test",
                ));
            }
        };

        let bytes = s.as_bytes();
        dst.reserve(bytes.len() + 1);

        for &b in bytes {
            // Client -> Server Encryption: Inverse of (b - 15) ^ 0xC3
            // b_enc = (b_dec ^ 0xC3) + 15
            dst.put_u8((b ^ 0xC3).wrapping_add(15));
        }
        dst.put_u8(0xD8); // Delimiter
        Ok(())
    }
}

impl Decoder for TestClientCodec {
    type Item = String; // Return raw string response
    type Error = io::Error;

    fn decode(&mut self, src: &mut BytesMut) -> Result<Option<Self::Item>, Self::Error> {
        // Server -> Client: Delimiter is 0x19
        if let Some(n) = src.iter().position(|&b| b == 0x19) {
            let data = src.split_to(n);
            src.advance(1); // skip 0x19

            let mut result = Vec::with_capacity(data.len());
            for &b in data.iter() {
                // Server -> Client Encryption was: b + 15
                // So Decryption is: b - 15
                result.push(b.wrapping_sub(15));
            }

            match String::from_utf8(result) {
                Ok(s) => Ok(Some(s)),
                Err(e) => Err(io::Error::new(io::ErrorKind::InvalidData, e)),
            }
        } else {
            Ok(None)
        }
    }
}

#[tokio::test]
async fn test_gateway_login_flow() -> Result<(), Box<dyn Error>> {
    let _ = env_logger::builder().is_test(true).try_init();

    // 1. Start server on a random port
    let listener = TcpListener::bind("127.0.0.1:0").await?;
    let addr = listener.local_addr()?;

    // Spawn server in a background task
    tokio::spawn(async move {
        if let Err(e) = elkia_gateway::run_server(listener).await {
            eprintln!("Server error: {}", e);
        }
    });

    // 2. Connect client
    let stream = TcpStream::connect(addr).await?;
    let mut client = Framed::new(stream, TestClientCodec);

    // 3. Send Login Command
    let login_cmd = LoginCommand {
        username: "ricofo8350@otanhome.com".to_string(),
        password: "9hibwiwiG2e6Nr".to_string(),
        client_version: "0.9.3+3086".to_string(),
    };

    client.send(AuthInteractRequest::Login(login_cmd)).await?;

    // 4. Expect Endpoint List Response
    if let Some(response) = client.next().await {
        let response = response?;
        println!("Received response: {}", response);

        // Verify response content
        // Expected: "NsTeST 0 127.0.0.1:5000:10:1.1.Elkia -1:-1:-1:10000.10000.1"
        assert!(response.starts_with("NsTeST 0 "));
        assert!(response.contains("127.0.0.1:5000"));
        assert!(response.contains("Elkia"));
    } else {
        panic!("Connection closed before receiving response");
    }

    Ok(())
}
