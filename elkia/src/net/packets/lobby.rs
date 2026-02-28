#[derive(Debug, PartialEq, Clone)]
pub enum LobbyCommandPacket {
    Select(SelectPacket),
    GameStart(GameStartPacket),
    CharNew(CharNewPacket),
}

#[derive(Debug, PartialEq, Clone)]
pub struct SelectPacket {
    pub slot: usize,
}

#[derive(Debug, PartialEq, Clone)]
pub struct GameStartPacket;

#[derive(Debug, PartialEq, Clone)]
pub struct CharNewPacket {
    pub name: String,
    pub slot: usize,
    pub class: i32,
}
