-- name: CreateCrowdStatus :one
INSERT INTO crowd_status (mess_id, level, updated_by) VALUES ($1, $2, $3) RETURNING *;

-- name: GetLatestCrowdStatus :one
SELECT * FROM crowd_status WHERE mess_id = $1 ORDER BY updated_at DESC LIMIT 1;
