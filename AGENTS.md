# Elkia - Project Overview for AI Agents

This document is intended to help AI agents understand the current state, architecture, and navigation of the Elkia project.

## Project Structure

The project has been refactored into a single Rust crate named `elkia` (formerly `elkia-core`, `elkia-net`, `elkia-gateway`, `elkia-world`).

- **`elkia/`**: Root directory of the Rust crate.
  - **`src/`**: Source code.
    - **`lib.rs`**: Library entry point, exports `auth`, `db`, `gateway`, `net`, `world`.
    - **`main.rs`**: Application entry point, handles CLI arguments and starts servers.
    - **`auth.rs`**: Authentication logic (`AuthService` trait, `SqliteAuthService`, `HandshakeData`).
    - **`db.rs`**: Database connection and migration logic (`sqlx` pool setup).
    - **`gateway/`**: The Auth/Gateway server logic (handles login and handoff).
    - **`net/`**: Networking code (packets, codecs, client handling).
      - **`codec/`**: `AuthCodec`, `GatewayCodec`, `HandshakeCodec`, `WorldCodec`.
      - **`packet/`**: Packet definitions (`handshake`, `gateway`, `lobby`, `world`, `game`).
    - **`world/`**: The World server logic (handles game loop, lobby, chat, movement).
      - **`state.rs`**: State machine implementation (`HandshakeState`, `LobbyState`, `GameState`).
    - **`lobby.rs`**: Lobby logic (`LobbyService` trait, `SqliteLobbyService`, character management).
    - **`game.rs`**: Game logic (`GameService` trait, `SqliteGameService`, movement, chat).
  - **`migrations/`**: SQLx migration files for SQLite.

## Key Architectural Decisions

1. **Single Monolithic Crate**: To simplify development and dependency management, all components are part of a single `elkia` crate.
2. **Gateway vs. World**:
   - **Gateway (Auth)**: Handles initial user authentication. Validates credentials (username/password) against the database and returns a session code + World server endpoint.
   - **World**: Handles the game session. Verifies the session code during handshake.
3. **Database**:
   - **SQLx with SQLite**: Primary data store.
   - **Integer IDs**: All primary keys (`id`) use `INTEGER PRIMARY KEY AUTOINCREMENT`.
   - **Schema**:
     - `accounts`: Stores username, password.
     - `sessions`: Stores session code, expiry, account association. Unique constraint on `(account_id, code)`.
     - `characters`: Stores character data (stats, appearance). Linked to `account_id`.
     - `account_characters`: Links accounts to characters with slot index.
     - `sessions_maps`: Stores the current/last position of a character (`map_id`, `position_x`, `position_y`).
4. **Networking**:
   - **Tokio**: Async I/O runtime.
   - **Codecs**: Custom `tokio_util::codec` implementations for packet framing.
   - **Packets**: Defined as Rust structs/enums, often mirroring legacy protocol.
5. **State Machine**:
   - The World server connection uses a state machine pattern: `HandshakeState` -> `LobbyState` -> `GameState`.

## Database Schema Details

- **IDs**: Use `i64` (Rust) / `INTEGER` (SQLite).
- **Position**: Character position is NOT stored in `characters` table. It is stored in `sessions_maps` to track the latest location per session/character.
- **Uniqueness**: `sessions` table enforces `UNIQUE(account_id, code)` to prevent code collisions for the same account.

## Authentication & Handshake Flow

1. **Gateway Login**: Client connects to Gateway -> Sends Login Packet (username, password) -> Gateway validates credentials -> Returns Session Code + World IP.
2. **World Handshake**:
   - Client connects to World Server.
   - Sends **Sync Command** (contains session code).
   - Sends **Username Command**.
   - Sends **Password Command** (contains password for verification).
   - Server performs **Transactional Verification** (`activate_session`):
     - **Credentials Check**: Verifies username/password against `accounts`.
     - **Active Session Check**: Ensures no other active session exists for this account (prevents double login).
     - **Session Code Check**: Verifies code matches and is not expired.
     - **Code Consumption**: Marks code as used (sets to NULL) if all checks pass.
   - **Outcome**:
     - If successful, transitions to **Lobby State**.
     - If failed, sends specific error packet:
       - `InvalidCredentials`: Wrong password or user not found.
       - `SessionAlreadyUsed`: User already logged in or session code reused.
       - `CannotAuthenticate`: Session expired or invalid code.
     - Closes connection on error.

3. **Session Cleanup**:
   - When a client disconnects, the World Server calls `AuthService::terminate_session(session_id)`.
   - This updates the session status to `terminated` (soft delete), preserving the record for history.

## Lobby & Character Management

- **Lobby Service**: Handles character list, creation, selection, deletion.
- **Selection**: When a character is selected, the system fetches the last known position from `sessions_maps` (or defaults to 1, 0, 0).
- **Creation**: Transactional creation of `characters` and `account_characters`.
- **Deletion**: Requires password verification (checked against `accounts` table using `WHERE id = ? AND password = ?`).

## CLI Usage

```bash
# Run Database Migrations
cargo run -p elkia -- migrate

# Run Gateway (Auth) Server
cargo run -p elkia -- gateway

# Run World Server
cargo run -p elkia -- world
```

## Development Notes

- **Refactoring**: When modifying service methods, prefer `account_id` over `username` for internal logic (World/Lobby/Game).
- **Transactions**: Use `sqlx::Transaction` for multi-step database operations (e.g., create character, select character).
- **Testing**: Use `cargo test`. Tests typically use an in-memory SQLite database (`sqlite::memory:`) and run migrations automatically.

## Legacy Mapping

- **Legacy `elkia-auth`** -> **`elkia gateway`**
- **Legacy `elkia-gateway`** -> **`elkia world`**
- **Legacy `internal/lobby`** -> **`elkia/src/lobby.rs`**
