-- name: CreateVideo :exec
INSERT INTO
    videos (
        id,
        upload_id,
        user_id,
        title,
        description,
        visibility,
        created_at
    )
VALUES
    (?, ?, ?, ?, ?, ?, unixepoch ("now"));

-- name: GetVideosFromSession :many
SELECT v.id, v.title, v.visibility, v.created_at
FROM sessions s
JOIN videos v ON s.user_id = v.user_id
WHERE s.token = ?;
