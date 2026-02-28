use async_trait::async_trait;
use sqlx::{Pool, Sqlite};
use std::error::Error;
use std::fmt;
use tracing::info;

#[derive(Debug)]
pub enum GameError {
    Sqlx(sqlx::Error),
}

impl fmt::Display for GameError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            GameError::Sqlx(e) => write!(f, "Database error: {}", e),
        }
    }
}

impl Error for GameError {
    fn source(&self) -> Option<&(dyn Error + 'static)> {
        match self {
            GameError::Sqlx(e) => Some(e),
        }
    }
}

impl From<sqlx::Error> for GameError {
    fn from(err: sqlx::Error) -> Self {
        GameError::Sqlx(err)
    }
}

#[async_trait]
pub trait GameService: Send + Sync {
    async fn walk(&self, session_id: i64, x: i32, y: i32) -> Result<(), GameError>;
    async fn chat(&self, session_id: i64, message: &str) -> Result<(), GameError>;
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
    async fn walk(&self, session_id: i64, x: i32, y: i32) -> Result<(), GameError> {
        info!("Character {} walked to ({}, {})", session_id, x, y);

        // Update session map position (if character is logged in)
        sqlx::query(
            r#"
            UPDATE sessions_maps
            SET position_x = ?, position_y = ?, last_seen = CURRENT_TIMESTAMP
            WHERE session_id = ?
            "#,
        )
        .bind(x)
        .bind(y)
        .bind(session_id)
        .execute(&self.pool)
        .await
        .map_err(GameError::from)?;

        Ok(())
    }

    async fn chat(&self, session_id: i64, message: &str) -> Result<(), GameError> {
        info!("Character {} says: {}", session_id, message);
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use sqlx::Row;

    #[tokio::test]
    async fn test_game_service() {
        let pool = sqlx::SqlitePool::connect("sqlite::memory:").await.unwrap();
        sqlx::migrate!("./migrations").run(&pool).await.unwrap();

        // Setup: Create user and character
        sqlx::query("INSERT INTO accounts (username, password) VALUES (?, ?)")
            .bind("test_user_game")
            .bind("password")
            .execute(&pool)
            .await
            .unwrap();

        let account_id: i64 = sqlx::query_scalar("SELECT id FROM accounts WHERE username = ?")
            .bind("test_user_game")
            .fetch_one(&pool)
            .await
            .unwrap();

        sqlx::query("INSERT INTO characters (account_id, name, class) VALUES (?, ?, ?)")
            .bind(account_id)
            .bind("Hero")
            .bind(1)
            .execute(&pool)
            .await
            .unwrap();

        let char_id: i64 = sqlx::query_scalar("SELECT id FROM characters WHERE name = ?")
            .bind("Hero")
            .fetch_one(&pool)
            .await
            .unwrap();

        // Setup: Create session and initial map position
        sqlx::query("INSERT INTO sessions (account_id, character_id, code) VALUES (?, ?, 12345)")
            .bind(account_id)
            .bind(char_id)
            .execute(&pool)
            .await
            .unwrap();

        let session_id: i64 = sqlx::query_scalar("SELECT id FROM sessions WHERE character_id = ?")
            .bind(char_id)
            .fetch_one(&pool)
            .await
            .unwrap();

        // Need maps and channels for foreign keys
        sqlx::query("INSERT INTO channels (name, port) VALUES ('ch1', 5000)")
            .execute(&pool)
            .await
            .unwrap();
        sqlx::query("INSERT INTO maps (name) VALUES ('map1')")
            .execute(&pool)
            .await
            .unwrap();

        sqlx::query(
            "INSERT INTO sessions_maps (session_id, channel_id, map_id, position_x, position_y) VALUES (?, 1, 1, 0, 0)",
        )
        .bind(session_id)
        .execute(&pool)
        .await
        .unwrap();

        let service = SqliteGameService::new(pool.clone());

        // Test walk
        service.walk(char_id, 10, 20).await.unwrap();

        // Verify position update in sessions_maps
        let row =
            sqlx::query("SELECT position_x, position_y FROM sessions_maps WHERE session_id = ?")
                .bind(session_id)
                .fetch_one(&pool)
                .await
                .unwrap();

        let x: i32 = row.get("position_x");
        let y: i32 = row.get("position_y");

        assert_eq!(x, 10);
        assert_eq!(y, 20);

        // Test chat (just ensures no error)
        service.chat(char_id, "Hello World").await.unwrap();
    }
}
