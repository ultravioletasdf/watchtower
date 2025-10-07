-- name: IsEmailTaken :one
SELECT
    1
FROM
    users
WHERE
    email
=
$1
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

-- name: GetUser :one
SELECT
    u.*,
    COALESCE(followers.count, 0) AS follower_count,
    COALESCE(following.count, 0) AS following_count
FROM users u
LEFT JOIN (
    SELECT user_id, COUNT(*) AS count
    FROM follows
    GROUP BY user_id
) followers ON followers.user_id = u.id
LEFT JOIN (
    SELECT follower_id, COUNT(*) AS count
    FROM follows
    GROUP BY follower_id
) following ON following.follower_id = u.id
WHERE username = $1;

-- name: GetUserById :one
SELECT
    u.*,
    COALESCE(followers.count, 0) AS follower_count,
    COALESCE(following.count, 0) AS following_count
FROM users u
LEFT JOIN (
    SELECT user_id, COUNT(*) AS count
    FROM follows
    GROUP BY user_id
) followers ON followers.user_id = u.id
LEFT JOIN (
    SELECT follower_id, COUNT(*) AS count
    FROM follows
    GROUP BY follower_id
) following ON following.follower_id = u.id
WHERE id = $1;

-- name: UpdateProfile :one
UPDATE users
SET display_name = COALESCE($1, display_name),
    description = COALESCE($2, description)
WHERE id = (
    SELECT user_id FROM sessions WHERE token = $3
)
RETURNING id;
