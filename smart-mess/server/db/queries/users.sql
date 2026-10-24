-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = lower(trim($1));

-- name: CreateUser :one
INSERT INTO users (email, role, hostel_group) VALUES (lower(trim($1)), $2, $3) RETURNING *;

-- name: SetHostelGroup :one
UPDATE users SET hostel_group = $2 WHERE id = $1 AND hostel_group IS NULL RETURNING *;

-- name: UpdatePreferences :one
UPDATE users SET diet_pref = COALESCE($2, diet_pref), cvd_mode = COALESCE($3, cvd_mode), text_scale = COALESCE($4, text_scale), theme = COALESCE($5, theme), reduced_motion = COALESCE($6, reduced_motion) WHERE id = $1 RETURNING *;

-- name: UpdateLastLogin :exec
UPDATE users SET last_login_at = now() WHERE id = $1;

-- name: GetManagerMesses :many
SELECT m.* FROM messes m JOIN manager_messes mm ON m.id = mm.mess_id WHERE mm.user_id = $1;
