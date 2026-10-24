-- name: CreateReview :one
INSERT INTO reviews (meal_id, user_id, variant, stars, comment) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: GetReviewByMealAndUser :one
SELECT * FROM reviews WHERE meal_id = $1 AND user_id = $2;

-- name: GetVisibleReviews :many
SELECT r.*, u.email as user_email FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.meal_id = $1 AND r.variant = $2 AND r.hidden = false ORDER BY r.created_at DESC LIMIT $3 OFFSET $4;

-- name: GetAllReviewsForMeal :many
SELECT r.*, u.email as user_email FROM reviews r JOIN users u ON r.user_id = u.id WHERE r.meal_id = $1 ORDER BY r.created_at DESC;

-- name: HideReview :exec
UPDATE reviews SET hidden = true WHERE id = $1;

-- name: UnhideReview :exec
UPDATE reviews SET hidden = false WHERE id = $1;

-- name: GetReviewByID :one
SELECT * FROM reviews WHERE id = $1;

-- name: GetMealRatingSummary :many
SELECT variant, COALESCE(AVG(stars), 0)::numeric(3,2) as avg_stars, COUNT(*) as review_count FROM reviews WHERE meal_id = $1 GROUP BY variant;
