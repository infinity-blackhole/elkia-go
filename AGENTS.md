# Elkia - Project Overview for AI Agents

This document is intended to help AI agents understand the current state, architecture, and navigation of the Elkia project.

## Project Structure

The project has been refactored into a single Rust crate named `elkia` (formerly `elkia-core`, `elkia-net`, `elkia-gateway`, `elkia-world`).

- **`elkia/`**: Root directory of the Rust crate.
  - **`src/`**: Source code.
    - **`lib.rs`**: Library entry point, exports `auth`, `db`, `gateway`, `net`, `world`.
    - **`main.rs`**: Application entry point, handles CLI arguments and starts servers.
    - **`auth/`**: Authentication logic (interfaces and implementations).
    - **`db/`**: Database connection and migration logic.
    - **`gateway/`**: The Auth/Gateway server logic (handles login and handoff).
    - **`net/`**: Networking code (packets, codecs, client handling). Moved from `elkia-net`.
    - **`world/`**: The World server logic (handles game loop, lobby, chat, movement).
  - **`migrations/`**: SQLx migration files for SQLite.

## Key Architectural Decisions

1. **Single Monolithic Crate**: To simplify development and dependency management, all components (gateway, world, net, core) are now part of a single `elkia` crate.
1. **Gateway vs. World**:
   - **Gateway (Auth)**: Handles initial user authentication. It validates credentials against the database and returns a session token and the address of the World server. It corresponds to the legacy `elkia-auth`.
   - **World**: Handles the actual game session, character selection, and gameplay. It verifies the session token with the database (or via shared secret/logic) upon connection. It corresponds to the legacy `elkia-gateway` (lobby) and game logic.
1. **Database**:
   - Switched from Redis to **SQLx with SQLite**.
   - Migrations are managed via `sqlx migrate`.
   - `elkia/src/db/` contains the connection pool setup.
1. **Networking**:
   - Uses `tokio` for async I/O.
   - `elkia-net` (now `elkia/src/net`) provides the `Codec` implementations (`AuthCodec`, `GatewayCodec`, `SessionCodec`) for framing TCP streams.
1. **Shared Logic**:
   - `elkia/src/auth/` defines the `AuthService` trait, implemented by `SqliteAuthService`.
   - Both Gateway and World servers use `AuthService` (Gateway for login, World for session verification).

## CLI Usage

The application uses `clap` for command-line arguments.

```bash
# Run Database Migrations
cargo run -p elkia -- migrate

# Run Gateway (Auth) Server
# Default: listens on 0.0.0.0:4000, hands off to 127.0.0.1:5000
cargo run -p elkia -- gateway

# Run World Server
# Default: listens on 0.0.0.0:5000
cargo run -p elkia -- world
```

## Legacy Mapping

- **Legacy `elkia-auth` (Go)** -> **`elkia gateway`** (Rust)
- **Legacy `elkia-gateway` (Go)** -> **`elkia world`** (Rust)
- **Legacy `internal/lobby`** -> **`elkia/src/world/services.rs`**

## Development Notes

- **Adding new packets**: Check `elkia/src/net/packets/`.
- **Modifying game logic**: Check `elkia/src/world/services.rs` (Lobby/Game services).
- **Database changes**: Create a new migration in `elkia/migrations/` and run `cargo run -p elkia -- migrate`.

## Current Status

- Basic Auth flow is implemented.
- World server accepts connections and verifies sessions.
- Character list and creation are implemented in-memory (mock).
- Networking stack is fully ported to Rust.
