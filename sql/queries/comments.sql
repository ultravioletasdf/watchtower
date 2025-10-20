-- name: ListComments :many
-- optimize at some point, three left joins is not necessary
SELECT
  c.*,
  u.username,
  count(l) AS likes,
  count(dl) AS dislikes,
  r.type
FROM
  comments c
  LEFT JOIN users u ON u.id = c.user_id
  -- like count
  LEFT JOIN reactions l ON c.id = l.target_id
  AND l.type = 1
  -- dislike count
  LEFT JOIN reactions dl ON c.id = dl.target_id
  AND dl.type = 2
  -- type of reaction for the user
  LEFT JOIN reactions r ON c.id = r.target_id
  AND r.user_id = $1
WHERE
  c.video_id = $2
  AND c.reference_id IS NOT DISTINCT FROM @reference_id -- filters out replies when @reference_id is set, list replies when it is
GROUP BY
  c.id,
  u.username,
  r.type
ORDER BY
  c.id DESC
LIMIT
  10
OFFSET
  $3;

-- name: CreateComment :exec
INSERT INTO comments (id, video_id, user_id, reference_id, content)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;

-- name: EditComment :exec
UPDATE comments
SET content = $1
WHERE id = $2;

-- name: ListReplies :many
SELECT * FROM comments
WHERE reference_id = $1
ORDER BY id DESC
LIMIT 10 OFFSET $2;
