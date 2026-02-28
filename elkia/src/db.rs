use sqlx::{migrate::MigrateDatabase, sqlite::SqlitePoolOptions, Sqlite, SqlitePool};
use std::error::Error;
use tracing::info;

pub async fn connect(url: &str) -> Result<SqlitePool, Box<dyn Error>> {
    if !Sqlite::database_exists(url).await.unwrap_or(false) {
        info!("Creating database {}", url);
        Sqlite::create_database(url).await?;
    }

    let pool = SqlitePoolOptions::new()
        .max_connections(5)
        .connect(url)
        .await?;

    Ok(pool)
}

pub async fn migrate(pool: &SqlitePool) -> Result<(), Box<dyn Error>> {
    info!("Running migrations...");
    sqlx::migrate!("./migrations").run(pool).await?;
    info!("Migrations complete.");
    Ok(())
}
