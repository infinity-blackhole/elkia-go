use async_trait::async_trait;
use serde::{Deserialize, Serialize};
use sqlx::{Pool, Row, Sqlite};
use std::collections::HashMap;
use std::sync::{Arc, Mutex};
use tracing::info;

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
    async fn create_character(
        &self,
        account_id: &str,
        name: &str,
        class: i32,
    ) -> Result<Character, String>;
}

#[async_trait]
pub trait GameService: Send + Sync {
    async fn walk(&self, char_id: &str, x: i32, y: i32);
    async fn chat(&self, char_id: &str, message: &str);
}

pub struct SqliteLobbyService {
    pool: Pool<Sqlite>,
}

impl SqliteLobbyService {
    pub fn new(pool: Pool<Sqlite>) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl LobbyService for SqliteLobbyService {
    async fn get_characters(&self, account_id: &str) -> Vec<Character> {
        let rows = sqlx::query("SELECT * FROM characters WHERE user_id = ?")
            .bind(account_id)
            .fetch_all(&self.pool)
            .await;

        match rows {
            Ok(rows) => rows
                .into_iter()
                .map(|row| Character {
                    id: row.get("id"),
                    name: row.get("name"),
                    class: row.get("class"),
                    level: row.get("level"),
                    hero_level: row.get("hero_level"),
                    job_level: row.get("job_level"),
                    experience: row.get("experience"),
                    job_experience: row.get("job_experience"),
                    hero_experience: row.get("hero_experience"),
                    faction: row.get("faction"),
                    reputation: row.get("reputation"),
                    dignity: row.get("dignity"),
                    compliment: row.get("compliment"),
                    health: row.get("health"),
                    mana: row.get("mana"),
                    hair_color: row.get("hair_color"),
                    hair_style: row.get("hair_style"),
                    x: row.get("x"),
                    y: row.get("y"),
                    map_id: row.get("map_id"),
                })
                .collect(),
            Err(e) => {
                tracing::error!("Failed to fetch characters: {}", e);
                Vec::new()
            }
        }
    }

    async fn create_character(
        &self,
        account_id: &str,
        name: &str,
        class: i32,
    ) -> Result<Character, String> {
        let id = uuid::Uuid::new_v4().to_string();
        let new_char = Character {
            id: id.clone(),
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

        sqlx::query(
            r#"
            INSERT INTO characters (
                id, user_id, name, class, level, hero_level, job_level, experience,
                job_experience, hero_experience, faction, reputation, dignity, compliment,
                health, mana, hair_color, hair_style, x, y, map_id
            )
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            "#,
        )
        .bind(&new_char.id)
        .bind(account_id)
        .bind(&new_char.name)
        .bind(new_char.class)
        .bind(new_char.level)
        .bind(new_char.hero_level)
        .bind(new_char.job_level)
        .bind(new_char.experience)
        .bind(new_char.job_experience)
        .bind(new_char.hero_experience)
        .bind(new_char.faction)
        .bind(new_char.reputation)
        .bind(new_char.dignity)
        .bind(new_char.compliment)
        .bind(new_char.health)
        .bind(new_char.mana)
        .bind(new_char.hair_color)
        .bind(new_char.hair_style)
        .bind(new_char.x)
        .bind(new_char.y)
        .bind(new_char.map_id)
        .execute(&self.pool)
        .await
        .map_err(|e| e.to_string())?;

        Ok(new_char)
    }
}

pub struct SqliteGameService {
    pool: Pool<Sqlite>,
}

impl SqliteGameService {
    pub fn new(pool: Pool<Sqlite>) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl GameService for SqliteGameService {
    async fn walk(&self, char_id: &str, x: i32, y: i32) {
        info!("Character {} walked to ({}, {})", char_id, x, y);
        if let Err(e) = sqlx::query("UPDATE characters SET x = ?, y = ? WHERE id = ?")
            .bind(x)
            .bind(y)
            .bind(char_id)
            .execute(&self.pool)
            .await
        {
            tracing::error!("Failed to update character position: {}", e);
        }
    }

    async fn chat(&self, char_id: &str, message: &str) {
        info!("Character {} says: {}", char_id, message);
    }
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

    async fn create_character(
        &self,
        account_id: &str,
        name: &str,
        class: i32,
    ) -> Result<Character, String> {
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
        info!("Character {} walked to ({}, {})", char_id, x, y);
    }

    async fn chat(&self, char_id: &str, message: &str) {
        info!("Character {} says: {}", char_id, message);
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

        let new_char = service
            .create_character(account_id, "Hero", 1)
            .await
            .unwrap();
        assert_eq!(new_char.name, "Hero");

        let chars = service.get_characters(account_id).await;
        assert_eq!(chars.len(), 1);
        assert_eq!(chars[0].name, "Hero");
    }
}
