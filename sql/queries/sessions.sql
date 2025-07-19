-- name: CreateSession :exec
INSERT INTO
    sessions (token, user_id, created_at)
VALUES
    (?, ?, unixepoch ('now'));
-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = ?;
-- name: GetUserFromSession :one
SELECT u.*
FROM sessions s
JOIN users u ON s.user_id = u.id
WHERE s.token = ?;