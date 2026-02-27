use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};

// Based on legacy/internal/lobby/lobby.go
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Character {
    pub id: String,
    pub name: String,
    pub class: i32,
    pub level: i32,
    pub hero_level: i32,
    pub job_level: i32,
    pub experience: i32,
    pub job_experience: i32,
    pub hero_experience: i32,
    pub faction: i32,
    pub reputation: i32,
    pub dignity: i32,
    pub compliment: i32,
    pub health: i32,
    pub mana: i32,
    pub hair_color: i32,
    pub hair_style: i32,
    pub x: i32,
    pub y: i32,
    pub map_id: i32,
}

#[async_trait]
pub trait LobbyService: Send + Sync {
    async fn get_characters(&self, account_id: &str) -> Vec<Character>;
    async fn create_character(&self, account_id: &str, name: &str, class: i32) -> Result<Character, String>;
}

#[async_trait]
pub trait GameService: Send + Sync {
    async fn walk(&self, char_id: &str, x: i32, y: i32);
    async fn chat(&self, char_id: &str, message: &str);
}

pub struct InMemoryLobbyService {
    characters: Arc<Mutex<HashMap<String, Vec<Character>>>>, // account_id -> characters
}

impl InMemoryLobbyService {
    pub fn new() -> Self {
        Self {
            characters: Arc::new(Mutex::new(HashMap::new())),
        }
    }
}

#[async_trait]
impl LobbyService for InMemoryLobbyService {
    async fn get_characters(&self, account_id: &str) -> Vec<Character> {
        let chars = self.characters.lock().unwrap();
        chars.get(account_id).cloned().unwrap_or_default()
    }

    async fn create_character(&self, account_id: &str, name: &str, class: i32) -> Result<Character, String> {
        let mut chars = self.characters.lock().unwrap();
        let user_chars = chars.entry(account_id.to_string()).or_insert_with(Vec::new);

        let new_char = Character {
            id: uuid::Uuid::new_v4().to_string(),
            name: name.to_string(),
            class,
            level: 1,
            hero_level: 0,
            job_level: 1,
            experience: 0,
            job_experience: 0,
            hero_experience: 0,
            faction: 0,
            reputation: 0,
            dignity: 100,
            compliment: 0,
            health: 100,
            mana: 100,
            hair_color: 0,
            hair_style: 0,
            x: 0,
            y: 0,
            map_id: 1,
        };

        user_chars.push(new_char.clone());
        Ok(new_char)
    }
}

pub struct InMemoryGameService {}

impl InMemoryGameService {
    pub fn new() -> Self {
        Self {}
    }
}

#[async_trait]
impl GameService for InMemoryGameService {
    async fn walk(&self, char_id: &str, x: i32, y: i32) {
        log::info!("Character {} walked to ({}, {})", char_id, x, y);
    }

    async fn chat(&self, char_id: &str, message: &str) {
        log::info!("Character {} says: {}", char_id, message);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn test_lobby_service() {
        let service = InMemoryLobbyService::new();
        let account_id = "test_acc";

        let chars = service.get_characters(account_id).await;
        assert!(chars.is_empty());

        let new_char = service.create_character(account_id, "Hero", 1).await.unwrap();
        assert_eq!(new_char.name, "Hero");

        let chars = service.get_characters(account_id).await;
        assert_eq!(chars.len(), 1);
        assert_eq!(chars[0].name, "Hero");
    }
}
