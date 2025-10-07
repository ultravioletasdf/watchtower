-- name: React :exec
INSERT INTO reactions (video_id, user_id, type)
VALUES ($1, (SELECT user_id FROM sessions WHERE token = $2), $3)
ON CONFLICT (video_id, user_id)
DO UPDATE SET
    type = EXCLUDED.type;

-- name: RemoveReaction :exec
DELETE FROM reactions
WHERE video_id = $1 AND user_id = (SELECT user_id FROM sessions WHERE token = $2);

-- name: GetReaction :one
SELECT type FROM reactions
WHERE video_id = $1 AND user_id = (SELECT user_id FROM sessions WHERE token = $2);
