-- name: CreateSession :exec
INSERT INTO
    sessions (token, user_id, created_at)
VALUES
    ($1, $2, CURRENT_TIMESTAMP);
-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = $1;
-- name: GetUserFromSession :one
SELECT u.*
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.token = $1;
