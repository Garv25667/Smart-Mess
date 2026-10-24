-- name: GetMealSchedules :many
SELECT * FROM meal_schedules WHERE hostel_group = $1 ORDER BY meal_type, day_kind;

-- name: CreateMealSchedule :one
INSERT INTO meal_schedules (hostel_group, meal_type, day_kind, start_time, end_time) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetPublishedMealsByDate :many
SELECT * FROM meals WHERE hostel_group = $1 AND service_date = $2 AND status = 'published' ORDER BY meal_type;

-- name: GetAllMealsByDate :many
SELECT * FROM meals WHERE hostel_group = $1 AND service_date = $2 ORDER BY meal_type;

-- name: GetMealByID :one
SELECT * FROM meals WHERE id = $1;

-- name: UpsertMeal :one
INSERT INTO meals (hostel_group, service_date, meal_type) VALUES ($1, $2, $3) ON CONFLICT (hostel_group, service_date, meal_type) DO UPDATE SET hostel_group = EXCLUDED.hostel_group RETURNING *;

-- name: PublishMeal :one
UPDATE meals SET status = 'published', published_at = now() WHERE id = $1 RETURNING *;

-- name: GetMealItems :many
SELECT mi.*, d.name as dish_name FROM meal_items mi JOIN dishes d ON mi.dish_id = d.id WHERE mi.meal_id = $1 ORDER BY mi.sort_order;

-- name: DeleteMealItems :exec
DELETE FROM meal_items WHERE meal_id = $1;

-- name: CreateMealItem :one
INSERT INTO meal_items (meal_id, dish_id, diet_menu, sort_order) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetScheduleFor :one
SELECT * FROM meal_schedules WHERE hostel_group = $1 AND meal_type = $2 AND day_kind = $3;
