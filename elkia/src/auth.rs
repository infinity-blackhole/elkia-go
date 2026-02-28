use async_trait::async_trait;
use chrono::{Duration, Utc};
use sqlx::{Pool, Row, Sqlite};
use tracing::{info, warn};
use uuid::Uuid;

#[derive(Debug, Clone)]
pub struct HandshakeData {
    pub id: String,
    pub user_id: String,
    pub username: String,
}

#[async_trait]
pub trait AuthService: Send + Sync {
    /// Authenticates user and creates a handshake flow, returning the code.
    async fn create_handshake_flow(&self, username: &str, password: &str) -> Result<u32, String>;

    /// Verifies the handshake ID and returns handshake data.
    async fn verify_handshake(&self, handshake_id: &str) -> Result<HandshakeData, String>;
}

pub struct SqliteAuthService {
    pool: Pool<Sqlite>,
}

impl SqliteAuthService {
    pub fn new(pool: Pool<Sqlite>) -> Self {
        Self { pool }
    }

    async fn verify_credentials(&self, username: &str, password: &str) -> Result<String, String> {
        let row = sqlx::query("SELECT id, password_hash FROM users WHERE username = ?")
            .bind(username)
            .fetch_optional(&self.pool)
            .await
            .map_err(|e| e.to_string())?;

        if let Some(row) = row {
            let user_id: String = row.try_get("id").map_err(|e| e.to_string())?;
            let password_hash: String = row.try_get("password_hash").map_err(|e| e.to_string())?;

            if password_hash != password {
                warn!("Invalid password for user: {}", username);
                return Err("Invalid credentials".to_string());
            }
            Ok(user_id)
        } else {
            warn!("User not found: {}", username);
            Err("User not found".to_string())
        }
    }
}

#[async_trait]
impl AuthService for SqliteAuthService {
    async fn create_handshake_flow(&self, username: &str, password: &str) -> Result<u32, String> {
        let user_id = self.verify_credentials(username, password).await?;
        let handshake_id = Uuid::new_v4().to_string();
        let code: u32 = rand::random();
        let expires_at = Utc::now() + Duration::hours(24);

        sqlx::query("INSERT INTO sessions (id, user_id, code, expires_at) VALUES (?, ?, ?, ?)")
            .bind(&handshake_id)
            .bind(&user_id)
            .bind(code)
            .bind(expires_at)
            .execute(&self.pool)
            .await
            .map_err(|e| e.to_string())?;

        info!(
            "Created handshake flow for user: {}, code: {}",
            username, code
        );
        Ok(code)
    }

    async fn verify_handshake(&self, handshake_id: &str) -> Result<HandshakeData, String> {
        let row = sqlx::query(
            r#"
            SELECT s.id, s.user_id, s.expires_at, u.username
            FROM sessions s
            JOIN users u ON s.user_id = u.id
            WHERE s.id = ?
            "#,
        )
        .bind(handshake_id)
        .fetch_optional(&self.pool)
        .await
        .map_err(|e| e.to_string())?;

        if let Some(row) = row {
            let expires_at: chrono::DateTime<Utc> =
                row.try_get("expires_at").map_err(|e| e.to_string())?;
            if expires_at < Utc::now() {
                warn!("Handshake expired: {}", handshake_id);
                return Err("Handshake expired".to_string());
            }

            let handshake_id: String = row.try_get("id").map_err(|e| e.to_string())?;
            let user_id: String = row.try_get("user_id").map_err(|e| e.to_string())?;
            let username: String = row.try_get("username").map_err(|e| e.to_string())?;

            Ok(HandshakeData {
                id: handshake_id,
                user_id,
                username,
            })
        } else {
            warn!("Handshake not found: {}", handshake_id);
            Err("Handshake not found".to_string())
        }
    }
}
