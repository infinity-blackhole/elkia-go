use async_trait::async_trait;
use chrono::{Duration, Utc};
use sqlx::{Pool, Row, Sqlite};
use std::error::Error;
use std::fmt;
use tracing::{info, warn};

#[derive(Debug)]
pub enum AuthError {
    InvalidCredentials,
    UserNotFound,
    HandshakeExpired,
    HandshakeNotFound,
    ActiveSession,
    DatabaseError(sqlx::Error),
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum SessionStatus {
    Active,
    Activating,
    Terminated,
}

impl fmt::Display for SessionStatus {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            SessionStatus::Active => write!(f, "active"),
            SessionStatus::Activating => write!(f, "activating"),
            SessionStatus::Terminated => write!(f, "terminated"),
        }
    }
}

impl fmt::Display for AuthError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            AuthError::InvalidCredentials => write!(f, "Invalid credentials"),
            AuthError::UserNotFound => write!(f, "User not found"),
            AuthError::HandshakeExpired => write!(f, "Handshake expired"),
            AuthError::HandshakeNotFound => write!(f, "Handshake not found"),
            AuthError::ActiveSession => write!(f, "Session already used"),
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

#[async_trait]
pub trait AuthService: Send + Sync {
    /// Authenticates user and creates a handshake flow, returning the code.
    async fn create_session(&self, username: &str, password: &str) -> Result<u32, AuthError>;

    /// Verifies the world login (credentials + handshake code + uniqueness check).
    async fn activate_session(
        &self,
        username: &str,
        password: &str,
        code: u32,
    ) -> Result<i64, AuthError>;

    /// Invalidates/Removes a session (logout).
    async fn terminate_session(&self, session_id: i64) -> Result<(), AuthError>;

    /// Updates the session's last_seen timestamp.
    async fn refresh_session(&self, session_id: i64) -> Result<(), AuthError>;
}

pub struct SqliteAuthService {
    pool: Pool<Sqlite>,
}

impl SqliteAuthService {
    pub fn new(pool: Pool<Sqlite>) -> Self {
        Self { pool }
    }
}

#[async_trait]
impl AuthService for SqliteAuthService {
    async fn create_session(&self, username: &str, password: &str) -> Result<u32, AuthError> {
        let mut tx = self.pool.begin().await.map_err(AuthError::DatabaseError)?;

        // 1. Verify Credentials
        let row = sqlx::query("SELECT id FROM accounts WHERE username = ? AND password = ?")
            .bind(username)
            .bind(password)
            .fetch_optional(&mut *tx)
            .await
            .map_err(AuthError::DatabaseError)?;

        let account_id = if let Some(row) = row {
            row.try_get::<i64, _>("id")
                .map_err(AuthError::DatabaseError)?
        } else {
            return Err(AuthError::InvalidCredentials);
        };

        // 2. Create Session
        let code: u32 = rand::random();
        let expires_at = Utc::now() + Duration::hours(24);

        sqlx::query(
            "INSERT INTO sessions (account_id, code, expires_at, status) VALUES (?, ?, ?, ?)",
        )
        .bind(account_id)
        .bind(code)
        .bind(expires_at)
        .bind(SessionStatus::Activating.to_string())
        .execute(&mut *tx)
        .await
        .map_err(AuthError::DatabaseError)?;

        tx.commit().await.map_err(AuthError::DatabaseError)?;

        info!(
            "Created handshake flow for user: {}, code: {}",
            username, code
        );
        Ok(code)
    }

    async fn activate_session(
        &self,
        username: &str,
        password: &str,
        code: u32,
    ) -> Result<i64, AuthError> {
        let mut tx = self.pool.begin().await.map_err(AuthError::DatabaseError)?;

        // 1. Verify Credentials
        let row = sqlx::query("SELECT id FROM accounts WHERE username = ? AND password = ?")
            .bind(username)
            .bind(password)
            .fetch_optional(&mut *tx)
            .await
            .map_err(AuthError::DatabaseError)?;

        let account_id = if let Some(row) = row {
            row.try_get::<i64, _>("id")
                .map_err(AuthError::DatabaseError)?
        } else {
            return Err(AuthError::InvalidCredentials);
        };

        // 2. Check for active session (already logged in)
        // We check for status = 'active'
        let active_session = sqlx::query(
            "SELECT id FROM sessions WHERE account_id = ? AND status = ? AND expires_at > ?",
        )
        .bind(account_id)
        .bind(SessionStatus::Active.to_string())
        .bind(Utc::now())
        .fetch_optional(&mut *tx)
        .await
        .map_err(AuthError::DatabaseError)?;

        if active_session.is_some() {
            warn!("User {} already has an active session", username);
            return Err(AuthError::ActiveSession);
        }

        // 3. Verify Handshake Code
        let session_row =
            sqlx::query("SELECT id, expires_at FROM sessions WHERE account_id = ? AND code = ?")
                .bind(account_id)
                .bind(code)
                .fetch_optional(&mut *tx)
                .await
                .map_err(AuthError::DatabaseError)?;

        let session_id = if let Some(row) = session_row {
            let expires_at: chrono::DateTime<Utc> = row
                .try_get("expires_at")
                .map_err(AuthError::DatabaseError)?;
            if expires_at < Utc::now() {
                warn!("Handshake expired for user: {}", username);
                return Err(AuthError::HandshakeExpired);
            }
            row.try_get::<i64, _>("id")
                .map_err(AuthError::DatabaseError)?
        } else {
            warn!("Handshake code not found or invalid for user: {}", username);
            return Err(AuthError::HandshakeNotFound);
        };

        // 4. Mark code as used (set to NULL) and status to Active
        sqlx::query("UPDATE sessions SET code = NULL, status = ? WHERE id = ?")
            .bind(SessionStatus::Active.to_string())
            .bind(session_id)
            .execute(&mut *tx)
            .await
            .map_err(AuthError::DatabaseError)?;

        tx.commit().await.map_err(AuthError::DatabaseError)?;

        Ok(session_id)
    }

    async fn terminate_session(&self, session_id: i64) -> Result<(), AuthError> {
        sqlx::query("UPDATE sessions SET status = ? WHERE id = ?")
            .bind(SessionStatus::Terminated.to_string())
            .bind(session_id)
            .execute(&self.pool)
            .await
            .map_err(AuthError::DatabaseError)?;
        Ok(())
    }

    async fn refresh_session(&self, session_id: i64) -> Result<(), AuthError> {
        sqlx::query("UPDATE sessions SET last_seen = CURRENT_TIMESTAMP WHERE id = ?")
            .bind(session_id)
            .execute(&self.pool)
            .await
            .map_err(AuthError::DatabaseError)?;
        Ok(())
    }
}
