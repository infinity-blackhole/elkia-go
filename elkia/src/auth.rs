use async_trait::async_trait;
use chrono::{Duration, Utc};
use sqlx::{Pool, Row, Sqlite};
use std::error::Error;
use std::fmt;
use tracing::{info, warn};
use uuid::Uuid;

#[derive(Debug)]
pub enum AuthError {
    InvalidCredentials,
    UserNotFound,
    HandshakeExpired,
    HandshakeNotFound,
    DatabaseError(sqlx::Error),
}

impl fmt::Display for AuthError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AuthError::InvalidCredentials => write!(f, "Invalid credentials"),
            AuthError::UserNotFound => write!(f, "User not found"),
            AuthError::HandshakeExpired => write!(f, "Handshake expired"),
            AuthError::HandshakeNotFound => write!(f, "Handshake not found"),
            AuthError::DatabaseError(e) => write!(f, "Database error: {}", e),
        }
    }
}

impl Error for AuthError {
    fn source(&self) -> Option<&(dyn Error + 'static)> {
        match self {
            AuthError::DatabaseError(e) => Some(e),
            _ => None,
        }
    }
}

impl From<sqlx::Error> for AuthError {
    fn from(err: sqlx::Error) -> Self {
        AuthError::DatabaseError(err)
    }
}

#[derive(Debug, Clone)]
pub struct HandshakeData {
    pub id: String,
    pub user_id: String,
    pub username: String,
}

#[async_trait]
pub trait AuthService: Send + Sync {
    /// Authenticates user and creates a handshake flow, returning the code.
    async fn create_handshake_flow(&self, username: &str, password: &str)
    -> Result<u32, AuthError>;

    /// Verifies the handshake ID and returns handshake data.
    async fn verify_handshake(&self, handshake_id: &str) -> Result<HandshakeData, AuthError>;
}

pub struct SqliteAuthService {
    pool: Pool<Sqlite>,
}

impl SqliteAuthService {
    pub fn new(pool: Pool<Sqlite>) -> Self {
        Self { pool }
    }

    async fn verify_credentials(
        &self,
        username: &str,
        password: &str,
    ) -> Result<String, AuthError> {
        let row = sqlx::query("SELECT id, password_hash FROM users WHERE username = ?")
            .bind(username)
            .fetch_optional(&self.pool)
            .await?;

        if let Some(row) = row {
            let user_id: String = row.try_get("id").map_err(AuthError::DatabaseError)?;
            let password_hash: String = row
                .try_get("password_hash")
                .map_err(AuthError::DatabaseError)?;

            if password_hash != password {
                warn!("Invalid password for user: {}", username);
                return Err(AuthError::InvalidCredentials);
            }
            Ok(user_id)
        } else {
            warn!("User not found: {}", username);
            Err(AuthError::UserNotFound)
        }
    }
}

#[async_trait]
impl AuthService for SqliteAuthService {
    async fn create_handshake_flow(
        &self,
        username: &str,
        password: &str,
    ) -> Result<u32, AuthError> {
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
            .await?;

        info!(
            "Created handshake flow for user: {}, code: {}",
            username, code
        );
        Ok(code)
    }

    async fn verify_handshake(&self, handshake_id: &str) -> Result<HandshakeData, AuthError> {
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
        .await?;

        if let Some(row) = row {
            let expires_at: chrono::DateTime<Utc> = row
                .try_get("expires_at")
                .map_err(AuthError::DatabaseError)?;
            if expires_at < Utc::now() {
                warn!("Handshake expired: {}", handshake_id);
                return Err(AuthError::HandshakeExpired);
            }

            let user_id: String = row.try_get("user_id").map_err(AuthError::DatabaseError)?;
            let username: String = row.try_get("username").map_err(AuthError::DatabaseError)?;

            Ok(HandshakeData {
                id: handshake_id.to_string(),
                user_id,
                username,
            })
        } else {
            warn!("Handshake not found: {}", handshake_id);
            Err(AuthError::HandshakeNotFound)
        }
    }
}
