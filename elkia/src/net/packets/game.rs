#[derive(Debug, PartialEq, Clone)]
pub enum GameCommandPacket {
    Walk(WalkPacket),
    Say(SayPacket),
}

#[derive(Debug, PartialEq, Clone)]
pub struct WalkPacket {
    pub x: i32,
    pub y: i32,
}

#[derive(Debug, PartialEq, Clone)]
pub struct SayPacket {
    pub message: String,
}
