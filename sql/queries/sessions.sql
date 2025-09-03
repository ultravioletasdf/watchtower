-- name: CreateSession :exec
INSERT INTO
    sessions (token, user_id)
VALUES
    ($1, $2);
-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = $1;
-- name: GetUserFromSession :one
SELECT
    u.*,
    COALESCE(followers.count, 0) AS follower_count,
    COALESCE(following.count, 0) AS following_count
FROM sessions s
JOIN users u ON s.user_id = u.id
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
WHERE s.token = $1;