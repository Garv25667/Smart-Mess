-- name: GetOrCreateDish :one
INSERT INTO dishes (hostel_group, name) VALUES ($1, $2) ON CONFLICT (hostel_group, lower(name)) DO UPDATE SET name = EXCLUDED.name RETURNING *;

-- name: GetDishByID :one
SELECT * FROM dishes WHERE id = $1;
