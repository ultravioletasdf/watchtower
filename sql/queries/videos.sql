-- name: CreateVideo :exec
INSERT INTO
    videos (
        id,
        upload_id,
        thumbnail_id,
        user_id,
        title,
        description,
        visibility,
        stage,
        created_at
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8, CURRENT_TIMESTAMP);

-- name: GetVideosFromSession :many
SELECT v.id, v.title, v.visibility, v.created_at, v.thumbnail_id, v.stage
FROM sessions s
JOIN videos v ON s.user_id = v.user_id
WHERE s.token = $1
ORDER BY v.created_at DESC;

-- name: GetUserVideos :many
SELECT id, title, visibility, created_at, thumbnail_id, stage
FROM videos
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetVideo :one
SELECT * FROM videos
WHERE id = $1;

-- name: DeleteVideo :exec
DELETE FROM videos
WHERE id = $1;

-- name: GetStage :one
SELECT stage, upload_id FROM videos
WHERE id = $1
LIMIT 1;

-- name: GetStages :one
SELECT json_object_agg(id, stage)
FROM videos
WHERE id = ANY(sqlc.arg(ids)::bigint[]);
