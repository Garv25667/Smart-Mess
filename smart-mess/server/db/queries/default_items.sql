-- name: GetDefaultItems :many
SELECT * FROM default_items WHERE hostel_group = $1 AND meal_type = $2 ORDER BY sort_order;

-- name: CreateDefaultItem :one
INSERT INTO default_items (hostel_group, meal_type, name, sort_order) VALUES ($1, $2, $3, $4) RETURNING *;
