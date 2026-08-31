-- name: GetLinks :many
SELECT id, original_url, short_name, created_at
FROM links
ORDER BY id;

-- name: GetLinkByID :one
SELECT id, original_url, short_name, created_at
FROM links
WHERE id = $1;

-- name: GetLinkByShortName :one
SELECT id, original_url, short_name, created_at
FROM links
WHERE short_name = $1;

-- name: CreateLink :one
INSERT INTO links (original_url, short_name)
VALUES ($1, $2)
RETURNING id, original_url, short_name, created_at;

-- name: UpdateLink :one
UPDATE links
SET original_url = $1, short_name = $2
WHERE id = $3
RETURNING id, original_url, short_name, created_at;

-- name: DeleteLink :exec
DELETE FROM links
WHERE id = $1;
