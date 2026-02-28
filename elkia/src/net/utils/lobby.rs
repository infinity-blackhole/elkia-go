use crate::net::codec::world::WorldCodec;
use crate::net::error::Error;
use crate::net::packet::lobby::{
    CharacterInfoPacket, CharacterListEndPacket, CharacterListStartPacket, LobbyEventPacket,
};
use crate::net::packet::world::WorldEventPacket;
use futures::stream::BoxStream;
use futures::{SinkExt, StreamExt};
use tokio::net::TcpStream;
use tokio_util::codec::Framed;

pub async fn send_character_list<'a>(
    framed: &mut Framed<TcpStream, WorldCodec>,
    mut stream: BoxStream<'a, CharacterInfoPacket>,
) -> Result<(), Error> {
    framed
        .send(WorldEventPacket::Lobby(
            LobbyEventPacket::CharacterListStart(CharacterListStartPacket { sequence: 0 }),
        ))
        .await?;

    while let Some(packet) = stream.next().await {
        framed
            .send(WorldEventPacket::Lobby(LobbyEventPacket::CharacterInfo(
                packet,
            )))
            .await?;
    }

    framed
        .send(WorldEventPacket::Lobby(LobbyEventPacket::CharacterListEnd(
            CharacterListEndPacket,
        )))
        .await?;

    Ok(())
}
