use async_trait::async_trait;
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Member {
    pub id: String,
    pub world_id: u32,
    pub channel_id: u32,
    pub name: String,
    pub addresses: Vec<String>,
    pub population: u32,
    pub capacity: u32,
}

#[async_trait]
pub trait CoordinatorService: Send + Sync {
    async fn member_add(&self, member: Member) -> Result<(), String>;
    async fn member_remove(&self, id: &str) -> Result<(), String>;
    async fn member_update(&self, id: &str, member: Member) -> Result<(), String>;
    async fn member_list(&self) -> Result<Vec<Member>, String>;
}

pub struct InMemoryCoordinatorService {
    members: std::sync::Arc<std::sync::Mutex<std::collections::HashMap<String, Member>>>,
}

impl InMemoryCoordinatorService {
    pub fn new() -> Self {
        Self {
            members: std::sync::Arc::new(std::sync::Mutex::new(std::collections::HashMap::new())),
        }
    }
}

#[async_trait]
impl CoordinatorService for InMemoryCoordinatorService {
    async fn member_add(&self, member: Member) -> Result<(), String> {
        let mut members = self.members.lock().unwrap();
        members.insert(member.id.clone(), member);
        Ok(())
    }

    async fn member_remove(&self, id: &str) -> Result<(), String> {
        let mut members = self.members.lock().unwrap();
        members.remove(id);
        Ok(())
    }

    async fn member_update(&self, id: &str, member: Member) -> Result<(), String> {
        let mut members = self.members.lock().unwrap();
        if members.contains_key(id) {
            members.insert(id.to_string(), member);
            Ok(())
        } else {
            Err("Member not found".to_string())
        }
    }

    async fn member_list(&self) -> Result<Vec<Member>, String> {
        let members = self.members.lock().unwrap();
        Ok(members.values().cloned().collect())
    }
}
