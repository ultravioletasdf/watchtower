-- name: CreateUpload :exec
INSERT INTO
    uploads (id, user_id, stage, created_at)
VALUES
    (?, ?, ?, ?);

-- name: GetUpload :one
SELECT
    *
FROM
    uploads
WHERE
    id = ?;

-- name: UpdateUploadStage :exec
UPDATE uploads
SET
    stage = ?
WHERE
    id = ?;