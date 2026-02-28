CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS characters (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    name TEXT NOT NULL UNIQUE,
    gender INTEGER NOT NULL DEFAULT 0,
    class INTEGER NOT NULL DEFAULT 1,
    level INTEGER NOT NULL DEFAULT 1,
    hero_level INTEGER NOT NULL DEFAULT 0,
    experience INTEGER NOT NULL DEFAULT 0,
    job_experience INTEGER NOT NULL DEFAULT 0,
    hero_experience INTEGER NOT NULL DEFAULT 0,
    job_level INTEGER NOT NULL DEFAULT 1,
    hair_color INTEGER NOT NULL DEFAULT 0,
    hair_style INTEGER NOT NULL DEFAULT 0,
    faction INTEGER NOT NULL DEFAULT 0,
    reputation INTEGER NOT NULL DEFAULT 0,
    dignity INTEGER NOT NULL DEFAULT 0,
    compliment INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(account_id) REFERENCES accounts(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sessions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id INTEGER NOT NULL,
    character_id INTEGER,
    code INTEGER,
    expires_at DATETIME,
    status TEXT NOT NULL DEFAULT 'activating',
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(account_id) REFERENCES accounts(id),
    FOREIGN KEY(character_id) REFERENCES characters(id) ON DELETE SET NULL,
    UNIQUE(account_id, code)
);

CREATE TABLE IF NOT EXISTS account_characters (
    account_id INTEGER NOT NULL,
    character_id INTEGER NOT NULL,
    slot INTEGER NOT NULL,
    PRIMARY KEY (account_id, slot),
    FOREIGN KEY(account_id) REFERENCES accounts(id),
    FOREIGN KEY(character_id) REFERENCES characters(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS channels (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    port INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS maps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS channel_maps (
    channel_id INTEGER NOT NULL,
    map_id INTEGER NOT NULL,
    PRIMARY KEY (channel_id, map_id),
    FOREIGN KEY(channel_id) REFERENCES channels(id),
    FOREIGN KEY(map_id) REFERENCES maps(id)
);

CREATE TABLE IF NOT EXISTS sessions_maps (
    session_id INTEGER PRIMARY KEY,
    channel_id INTEGER NOT NULL,
    map_id INTEGER NOT NULL,
    position_x INTEGER NOT NULL DEFAULT 0,
    position_y INTEGER NOT NULL DEFAULT 0,
    last_seen DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    FOREIGN KEY(channel_id) REFERENCES channels(id),
    FOREIGN KEY(map_id) REFERENCES maps(id)
);
