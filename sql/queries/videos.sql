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