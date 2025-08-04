-- name: IsEmailTaken :one
SELECT
    1
FROM
    users
WHERE
    email = $1
LIMIT
    1;

-- name: IsUsernameTaken :one
SELECT
    1
FROM
    users
WHERE
    username = $1
LIMIT
    1;

-- name: CreateUser :exec
INSERT INTO
    users (id, email, username, password, verify_code, verify_expire_at, created_at)
VALUES
    ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP);

-- name: GetPasswordFromEmail :one
SELECT
    id, password
FROM
    users
WHERE
    email = $1
LIMIT
    1;

-- name: SetUserFlag :exec
UPDATE users
SET flags = $1
WHERE id = $2;
