use crate::net::packet::lobby::CharacterInfoPacket;
use async_trait::async_trait;
use futures::stream::BoxStream;
use futures::{StreamExt, TryStreamExt};
use serde::{Deserialize, Serialize};
use sqlx::{Pool, Sqlite};
use std::error::Error;
use std::fmt;

#[derive(Debug)]
pub enum LobbyError {
    Sqlx(sqlx::Error),
}

impl fmt::Display for LobbyError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            LobbyError::Sqlx(e) => write!(f, "Database error: {}", e),
        }
    }
}

impl Error for LobbyError {
    fn source(&self) -> Option<&(dyn Error + 'static)> {
        match self {
            LobbyError::Sqlx(e) => Some(e),
        }
    }
}

impl From<sqlx::Error> for LobbyError {
    fn from(err: sqlx::Error) -> Self {
        LobbyError::Sqlx(err)
    }
}

// Based on legacy/internal/lobby/lobby.go
#[derive(Debug, Clone, Serialize, Deserialize, sqlx::FromRow)]
pub struct Character {
    pub id: i64,
    pub name: String,
    pub slot: i32,
    pub gender: i32,
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
    pub hair_color: i32,
    pub hair_style: i32,
}

impl From<Character> for CharacterInfoPacket {
    fn from(c: Character) -> Self {
        CharacterInfoPacket {
            name: c.name,
            id: c.id,
            slot: c.slot,
            gender: c.gender,
            class: c.class,
            level: c.level,
            hero_level: c.hero_level,
            hair_color: c.hair_color,
            hair_style: c.hair_style,
            faction: c.faction,
            reputation: c.reputation,
            dignity: c.dignity,
            compliment: c.compliment,
            job_level: c.job_level,
            experience: c.experience,
            job_experience: c.job_experience,
            hero_experience: c.hero_experience,
        }
    }
}

#[async_trait]
pub trait LobbyService: Send + Sync {
    fn list_characters<'a>(
        &'a self,
        session_id: i64,
    ) -> BoxStream<'a, Result<Character, LobbyError>>;
    fn create_character<'a>(
        &'a self,
        session_id: i64,
        name: &'a str,
        slot: i32,
        gender: i32,
        hair_style: i32,
        hair_color: i32,
    ) -> BoxStream<'a, Result<Character, LobbyError>>;
    fn get_character<'a>(&'a self, char_id: i64) -> BoxStream<'a, Result<Character, LobbyError>>;
    async fn select_character(&self, session_id: i64, slot: usize) -> Result<i64, LobbyError>;
    fn delete_character<'a>(
        &'a self,
        session_id: i64,
        slot: usize,
        password: &'a str,
    ) -> BoxStream<'a, Result<Character, LobbyError>>;
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
    fn list_characters<'a>(
        &'a self,
        session_id: i64,
    ) -> BoxStream<'a, Result<Character, LobbyError>> {
        sqlx::query_as::<_, Character>(
            r#"
            SELECT c.*, ac.slot
            FROM characters c
            JOIN account_characters ac ON c.id = ac.character_id
            JOIN sessions s ON s.account_id = c.account_id
            WHERE s.id = ?
            ORDER BY ac.slot ASC
            "#,
        )
        .bind(session_id)
        .fetch(&self.pool)
        .map(|res| res.map_err(LobbyError::from))
        .boxed()
    }

    fn create_character<'a>(
        &'a self,
        session_id: i64,
        name: &'a str,
        slot: i32,
        gender: i32,
        hair_style: i32,
        hair_color: i32,
    ) -> BoxStream<'a, Result<Character, LobbyError>> {
        sqlx::query_as::<_, Character>(
            r#"
            INSERT INTO characters (account_id, name, gender, hair_style, hair_color)
            SELECT s.account_id, ?, ?, ?, ? FROM sessions s WHERE s.id = ?;

            INSERT INTO account_characters (account_id, character_id, slot)
            SELECT s.account_id, last_insert_rowid(), ? FROM sessions s WHERE s.id = ?;

            SELECT c.*, ac.slot
            FROM characters c
            JOIN account_characters ac ON c.id = ac.character_id
            JOIN sessions s ON s.account_id = c.account_id
            WHERE s.id = ?
            ORDER BY ac.slot ASC
            "#,
        )
        .bind(name.to_string())
        .bind(gender)
        .bind(hair_style)
        .bind(hair_color)
        .bind(session_id)
        .bind(slot)
        .bind(session_id)
        .bind(session_id)
        .fetch(&self.pool)
        .map(|res| res.map_err(LobbyError::from))
        .boxed()
    }

    fn get_character<'a>(&'a self, char_id: i64) -> BoxStream<'a, Result<Character, LobbyError>> {
        sqlx::query_as::<_, Character>(
            r#"
            SELECT c.*, ac.slot
            FROM characters c
            JOIN account_characters ac ON c.id = ac.character_id
            WHERE c.id = ?
            "#,
        )
        .bind(char_id)
        .fetch(&self.pool)
        .map(|res| res.map_err(LobbyError::from))
        .boxed()
    }

    async fn select_character(&self, session_id: i64, slot: usize) -> Result<i64, LobbyError> {
        let mut tx = self.pool.begin().await.map_err(LobbyError::from)?;

        // 1. Get character info
        let row: Option<(i64,)> = sqlx::query_as(
            r#"
            SELECT c.id
            FROM account_characters ac
            JOIN characters c ON ac.character_id = c.id
            JOIN sessions s ON s.account_id = ac.account_id
            WHERE s.id = ? AND ac.slot = ?
            "#,
        )
        .bind(session_id)
        .bind(slot as i32)
        .fetch_optional(&mut *tx)
        .await
        .map_err(LobbyError::from)?;

        let (char_id,) = row
            .ok_or(sqlx::Error::RowNotFound)
            .map_err(LobbyError::from)?;

        // 2. Get last known position
        let pos_row: Option<(i32, i32, i32)> = sqlx::query_as(
            r#"
            SELECT sm.map_id, sm.position_x, sm.position_y
            FROM sessions s
            JOIN sessions_maps sm ON s.id = sm.session_id
            WHERE s.character_id = ?
            ORDER BY sm.last_seen DESC
            LIMIT 1
            "#,
        )
        .bind(char_id)
        .fetch_optional(&mut *tx)
        .await
        .map_err(LobbyError::from)?;

        let (map_id, x, y) = pos_row.unwrap_or((1, 0, 0)); // Default spawn

        // 3. Update sessions
        sqlx::query("UPDATE sessions SET character_id = ? WHERE id = ?")
            .bind(char_id)
            .bind(session_id)
            .execute(&mut *tx)
            .await
            .map_err(LobbyError::from)?;

        // 4. Insert into sessions_maps
        sqlx::query(
            r#"
            INSERT OR REPLACE INTO sessions_maps (session_id, channel_id, map_id, position_x, position_y)
            VALUES (?, 1, ?, ?, ?)
            "#,
        )
        .bind(session_id)
        .bind(map_id)
        .bind(x)
        .bind(y)
        .execute(&mut *tx)
        .await
        .map_err(LobbyError::from)?;

        tx.commit().await.map_err(LobbyError::from)?;

        Ok(char_id)
    }

    fn delete_character<'a>(
        &'a self,
        session_id: i64,
        slot: usize,
        password: &'a str,
    ) -> BoxStream<'a, Result<Character, LobbyError>> {
        let password = password.to_string();
        futures::stream::once(async move {
            let mut tx = self.pool.begin().await.map_err(LobbyError::from)?;

            // 1. Verify password using JOIN
            let auth_check: Option<(i64,)> = sqlx::query_as(
                r#"
                SELECT a.id
                FROM accounts a
                JOIN sessions s ON s.account_id = a.id
                WHERE s.id = ? AND a.password = ?
                "#,
            )
            .bind(session_id)
            .bind(&password)
            .fetch_optional(&mut *tx)
            .await
            .map_err(LobbyError::from)?;

            if auth_check.is_none() {
                return Err(LobbyError::Sqlx(sqlx::Error::RowNotFound));
            }

            // 2. Get character_id using JOIN
            let row: Option<(i64,)> = sqlx::query_as(
                r#"
                SELECT ac.character_id
                FROM account_characters ac
                JOIN sessions s ON s.account_id = ac.account_id
                WHERE s.id = ? AND ac.slot = ?
                "#,
            )
            .bind(session_id)
            .bind(slot as i32)
            .fetch_optional(&mut *tx)
            .await
            .map_err(LobbyError::from)?;

            if let Some((char_id,)) = row {
                // 3. Delete character
                sqlx::query("DELETE FROM characters WHERE id = ?")
                    .bind(char_id)
                    .execute(&mut *tx)
                    .await
                    .map_err(LobbyError::from)?;
            } else {
                return Err(LobbyError::Sqlx(sqlx::Error::RowNotFound));
            }

            tx.commit().await.map_err(LobbyError::from)?;
            Ok(session_id)
        })
        .map_ok(
            move |session_id| -> BoxStream<'a, Result<Character, LobbyError>> {
                self.list_characters(session_id)
            },
        )
        .try_flatten()
        .boxed()
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::db;

    #[tokio::test]
    async fn test_lobby_service() {
        let pool = db::connect("sqlite::memory:").await.unwrap();
        // Assuming migrations are run or schema is set up.
        // For testing, we might need to run migration manually if db::migrate reads from file system
        // and we are in a tool environment where paths might be tricky.
        // But assuming db::migrate works.
        db::migrate(&pool).await.unwrap();

        // Insert default channel and map for testing (required by select_character)
        sqlx::query("INSERT INTO channels (id, name, port) VALUES (1, 'Channel 1', 5000)")
            .execute(&pool)
            .await
            .unwrap();
        sqlx::query("INSERT INTO maps (id, name) VALUES (1, 'Map 1')")
            .execute(&pool)
            .await
            .unwrap();

        let username = "test_user_2";
        // Insert test user
        sqlx::query("INSERT INTO accounts (username, password) VALUES (?, ?)")
            .bind(username)
            .bind("password")
            .execute(&pool)
            .await
            .unwrap();

        let account_id: i64 = sqlx::query_scalar("SELECT id FROM accounts WHERE username = ?")
            .bind(username)
            .fetch_one(&pool)
            .await
            .unwrap();

        // Insert test session
        sqlx::query("INSERT INTO sessions (account_id, code) VALUES (?, ?)")
            .bind(account_id)
            .bind(12345)
            .execute(&pool)
            .await
            .unwrap();

        let session_id: i64 = sqlx::query_scalar("SELECT id FROM sessions WHERE account_id = ?")
            .bind(account_id)
            .fetch_one(&pool)
            .await
            .unwrap();

        let service = SqliteLobbyService::new(pool.clone());

        // Test create_character
        let mut stream = service.create_character(session_id, "Hero", 0, 1, 1, 2);
        let char_res = stream.next().await.unwrap().unwrap();
        assert_eq!(char_res.name, "Hero");
        assert_eq!(char_res.slot, 0);
        assert_eq!(char_res.gender, 1);

        // Test list_characters
        let mut stream = service.list_characters(session_id);
        let char_res = stream.next().await.unwrap().unwrap();
        assert_eq!(char_res.name, "Hero");

        // Test select_character
        let char_id = service.select_character(session_id, 0).await.unwrap();
        assert_eq!(char_id, char_res.id);

        // Test delete_character
        let mut stream = service.delete_character(session_id, 0, "password");
        let res = stream.next().await;
        // Should be empty or None if list is empty?
        // list_characters returns a stream. If empty, next() is None.
        assert!(res.is_none());
    }
}
