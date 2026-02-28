use sqlx::{Sqlite, SqlitePool, migrate::MigrateDatabase, sqlite::SqlitePoolOptions};
use std::error::Error;
use tracing::info;

pub async fn connect() -> Result<SqlitePool, Box<dyn Error>> {
    let database_url =
        std::env::var("DATABASE_URL").unwrap_or_else(|_| "sqlite:elkia.db".to_string());

    if !Sqlite::database_exists(&database_url)
        .await
        .unwrap_or(false)
    {
        info!("Creating database {}", database_url);
        Sqlite::create_database(&database_url).await?;
    }

    let pool = SqlitePoolOptions::new()
        .max_connections(5)
        .connect(&database_url)
        .await?;

    Ok(pool)
}

pub async fn migrate(pool: &SqlitePool) -> Result<(), Box<dyn Error>> {
    info!("Running migrations...");
    sqlx::migrate!("./migrations").run(pool).await?;
    info!("Migrations complete.");
    Ok(())
}
