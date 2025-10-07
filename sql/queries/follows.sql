-- name: FollowUser :exec
WITH session AS (
	SELECT s.user_id
	FROM sessions s
	WHERE s.token = $1
)
INSERT INTO follows (user_id, follower_id)
SELECT $2, session.user_id
FROM session;

-- name: UnfollowUser :exec
WITH session AS (
	SELECT s.user_id
	FROM sessions s
	WHERE s.token = $1
)
DELETE FROM follows f
USING session
WHERE f.user_id = $2
AND f.follower_id = session.user_id;

-- name: CountFollows :one
SELECT count(*) FROM follows WHERE user_id = $1;

-- name: GetUserFollows :many
SELECT f.user_id, f.created_at, u.username FROM follows f
LEFT JOIN users u ON f.user_id = u.id
WHERE f.follower_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- name: GetUserFollowers :many
SELECT f.follower_id, f.created_at, u.username FROM follows f
LEFT JOIN users u ON f.follower_id = u.id
WHERE f.user_id = $1
ORDER BY f.created_at DESC
LIMIT $2 OFFSET $3;

-- name: IsFollowing :one
SELECT 1 FROM follows f
JOIN sessions s ON s.user_id = f.follower_id
WHERE f.user_id = $1 AND s.token = $2;
