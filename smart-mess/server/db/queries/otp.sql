-- name: CreateOTP :one
INSERT INTO otp_codes (email, code_hash, expires_at) VALUES (lower(trim($1)), $2, $3) RETURNING *;

-- name: GetActiveOTP :one
SELECT * FROM otp_codes WHERE email = lower(trim($1)) AND consumed_at IS NULL AND expires_at > now() ORDER BY created_at DESC LIMIT 1;

-- name: IncrementOTPAttempts :exec
UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1;

-- name: ConsumeOTP :exec
UPDATE otp_codes SET consumed_at = now() WHERE id = $1;

-- name: CountRecentOTPs :one
SELECT count(*) FROM otp_codes WHERE email = lower(trim($1)) AND created_at > now() - interval '1 hour';
