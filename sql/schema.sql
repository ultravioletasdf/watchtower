CREATE TABLE
users (
    id bigint PRIMARY KEY,
    email text NOT NULL,
    username varchar(72) NOT NULL,
    password bytea NOT NULL,
    created_at timestamptz NOT NULL,
    flags int NOT NULL DEFAULT 0,
    verify_code int NOT NULL DEFAULT 0,
    verify_expire_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX idx_users_email ON users (email);
CREATE UNIQUE INDEX idx_users_username ON users (username);

CREATE TABLE
sessions (
    token char(32) PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE
uploads (
    id bigint PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL
);

CREATE TABLE
videos (
    id bigint PRIMARY KEY,
    upload_id bigint NOT NULL,
    thumbnail_id bigint NOT NULL REFERENCES uploads (id),
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    title varchar(100) NOT NULL,
    description varchar(1000) NOT NULL,
    visibility int NOT NULL DEFAULT 0,
    stage int NOT NULL,
    created_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE TABLE follows (
    user_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    follower_id bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, follower_id)
);
