PRAGMA foreign_keys = on;

CREATE TABLE
    users (
        id INTEGER PRIMARY KEY,
        email TEXT NOT NULL,
        username TEXT NOT NULL,
        password BLOB NOT NULL,
        created_at INTEGER NOT NULL,
        flags INTEGER NOT NULL DEFAULT 0,
        verify_code INTEGER NOT NULL DEFAULT 0,
        verify_expire_at INTEGER NOT NULL DEFAULT 0
    );

CREATE UNIQUE INDEX idx_users_email ON users (email);

CREATE UNIQUE INDEX idx_users_username ON users (username);

CREATE TABLE
    sessions (
        token TEXT PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        created_at INTEGER NOT NULL
    );

CREATE TABLE
    uploads (
        id INTEGER PRIMARY KEY,
        user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        stage INTEGER NOT NULL,
        created_at INTEGER NOT NULL
    );

CREATE TABLE
    videos (
        id INTEGER PRIMARY KEY,
        upload_id INTEGER NOT NULL REFERENCES uploads (id),
        user_id INTEGER NOT NULL REFERENCES users (id) ON DELETE CASCADE,
        title TEXT NOT NULL,
        description TEXT NOT NULL,
        visibility INTEGER NOT NULL DEFAULT 0,
        created_at INTEGER NOT NULL
    );