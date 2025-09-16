-- name: CreateUpload :exec
INSERT INTO
    uploads (id, user_id, created_at)
VALUES
    ($1, $2, CURRENT_TIMESTAMP);

-- name: GetUpload :one
SELECT
    *
FROM
    uploads
WHERE
    id = $1;
