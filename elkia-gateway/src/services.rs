use async_trait::async_trait;
use std::sync::{Arc, Mutex};
use std::collections::HashMap;

#[async_trait]
pub trait AuthService: Send + Sync {
    async fn login(&self, username: &str, password: &str) -> Result<AuthResult, String>;
}

#[derive(Debug, Clone)]
pub struct AuthResult {
    pub session_id: String,
    pub code: u32,
}

pub struct InMemoryAuthService {
    // In a real app, this would be a DB connection or similar
    sessions: Arc<Mutex<HashMap<String, AuthResult>>>,
}

impl InMemoryAuthService {
    pub fn new() -> Self {
        Self {
            sessions: Arc::new(Mutex::new(HashMap::new())),
        }
    }
}

#[async_trait]
impl AuthService for InMemoryAuthService {
    async fn login(&self, username: &str, _password: &str) -> Result<AuthResult, String> {
        // Mock authentication: Accept any user, generate a session
        let session_id = uuid::Uuid::new_v4().to_string();
        let code = rand::random::<u32>(); // Random code for session encryption

        let result = AuthResult {
            session_id: session_id.clone(),
            code,
        };

        let mut sessions = self.sessions.lock().unwrap();
        sessions.insert(username.to_string(), result.clone());

        Ok(result)
    }
}
