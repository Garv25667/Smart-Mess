-- name: ListMessesByGroup :many
SELECT * FROM messes WHERE hostel_group = $1 ORDER BY name;

-- name: GetMessByID :one
SELECT * FROM messes WHERE id = $1;

-- name: GetMessByCode :one
SELECT * FROM messes WHERE code = $1;

-- name: CreateMess :one
INSERT INTO messes (code, name, hostel_group) VALUES ($1, $2, $3) RETURNING *;

-- name: IsMessAssignedToManager :one
SELECT EXISTS(SELECT 1 FROM manager_messes WHERE user_id = $1 AND mess_id = $2);

-- name: AssignMessToManager :exec
INSERT INTO manager_messes (user_id, mess_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;
