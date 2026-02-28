use sqlx::sqlite::SqliteConnectOptions;
use std::str::FromStr;

fn main() {
    let _opts = SqliteConnectOptions::from_str("sqlite::memory:").unwrap();
    // check if we can call something like multiple_statements or similar
    // _opts.multiple_statements(true); // Uncomment to test
}
