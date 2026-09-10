-- name: CreateVisit :execrows
INSERT INTO link_visits (link_id, ip, referer, user_agent, status)
VALUES ($1, $2, $3, $4, $5);

-- name: GetLinkVisits :many
SELECT id, link_id, ip, referer, user_agent, status, created_at
FROM link_visits
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountLinkVisits :one
SELECT COUNT(*)::bigint AS count FROM link_visits;
