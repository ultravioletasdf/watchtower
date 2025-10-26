-- name: React :exec
INSERT INTO reactions (target_id, user_id, type)
VALUES ($1, $2, $3)
ON CONFLICT (target_id, user_id) DO UPDATE SET
    type = EXCLUDED.type;

-- name: RemoveReaction :exec
DELETE FROM reactions
WHERE target_id = $1 AND user_id = $2;

-- name: GetReaction :one
SELECT type FROM reactions
WHERE target_id = $1 AND user_id = $2;
