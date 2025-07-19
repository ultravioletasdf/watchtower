-- name: IsEmailTaken :one
SELECT
    1
FROM
    users
WHERE
    email = ?
LIMIT
    1;

-- name: IsUsernameTaken :one
SELECT
    1
FROM
    users
WHERE
    username = ?
LIMIT
    1;

-- name: CreateUser :exec
INSERT INTO
    users (id, email, username, password, verify_code, verify_expire_at, created_at)
VALUES
    (?, ?, ?, ?, ?, ?, unixepoch ('now'));

-- name: GetPasswordFromEmail :one
SELECT
    id, password
FROM
    users
WHERE
    email = ?
LIMIT
    1;

-- name: SetUserFlag :exec
UPDATE users
SET flags = ?
WHERE id = ?;