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
        stage
    )
VALUES
    ($1, $2, $3, $4, $5, $6, $7, $8);

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
    AND (@show_private::BOOL OR visibility = 0)
ORDER BY created_at DESC;

-- name: GetVideo :one
SELECT
    v.*,
    COALESCE(likes.count, 0) AS likes,
    COALESCE(dislikes.count, 0) AS dislikes,
    COALESCE(comments.count, 0) AS comments
FROM videos v
LEFT JOIN (
    SELECT target_id, COUNT(*) AS count
    FROM reactions
    WHERE type = 1
    GROUP BY target_id
) AS likes ON likes.target_id = v.id
LEFT JOIN (
    SELECT target_id, COUNT(*) AS count
    FROM reactions
    WHERE type = 2
    GROUP BY target_id
) AS dislikes ON dislikes.target_id = v.id
LEFT JOIN LATERAL (
  SELECT COUNT(*) AS count
  FROM comments c
  WHERE c.video_id = $1 AND c.reference_id IS NULL
) AS comments ON true
WHERE v.id = $1;

-- name: GetVideoByUploadId :one
SELECT * FROM videos
WHERE upload_id = $1;

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

-- name: GetUsersFollowingVideos :many
SELECT v.*
FROM videos as v
JOIN follows as f
    ON v.user_id = f.user_id
WHERE f.follower_id = $1
    AND v.stage = 3
    AND visibility = 0
ORDER BY v.id DESC
LIMIT 10
OFFSET $2;

-- name: UpdateVideoStage :exec
UPDATE videos
SET
    stage = $1
WHERE
    id = $2
    AND STAGE <> 4;
