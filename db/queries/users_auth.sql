-- ───────────────────────── Users ─────────────────────────

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: ListUsersByRole :many
SELECT * FROM users
WHERE role = $1
ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUserPasswordHash :one
UPDATE users SET password_hash = $2 WHERE id = $1 RETURNING *;

-- name: UpdateUserRole :one
UPDATE users SET role = $2 WHERE id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- ───────────────────────── Passkeys ─────────────────────────

-- name: CreatePasskey :one
INSERT INTO passkeys (user_id, credential_id, public_key, sign_count)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetPasskeyByCredentialID :one
SELECT * FROM passkeys WHERE credential_id = $1;

-- name: ListPasskeysByUserID :many
SELECT * FROM passkeys WHERE user_id = $1 ORDER BY created_at DESC;

-- name: UpdatePasskeySignCount :one
UPDATE passkeys SET sign_count = $2 WHERE id = $1 RETURNING *;

-- name: DeletePasskey :exec
DELETE FROM passkeys WHERE id = $1;

-- ───────────────────────── Sessions ─────────────────────────

-- name: CreateSession :one
INSERT INTO sessions (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionByTokenHash :one
SELECT * FROM sessions WHERE token_hash = $1;

-- name: UpdateSessionLastSeen :exec
UPDATE sessions SET last_seen_at = NOW() WHERE id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < NOW();

-- ───────────────────────── Login Attempts ─────────────────────────

-- name: CreateLoginAttempt :one
INSERT INTO login_attempts (username, ip, password_fingerprint, success)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListRecentLoginAttemptsByUsername :many
SELECT * FROM login_attempts
WHERE username = $1
  AND attempted_at > NOW() - INTERVAL '1 hour'
ORDER BY attempted_at DESC;

-- name: ListRecentLoginAttemptsByIP :many
SELECT * FROM login_attempts
WHERE ip = $1
  AND attempted_at > NOW() - INTERVAL '1 hour'
ORDER BY attempted_at DESC;

-- ───────────────────────── Login Rate Limits ─────────────────────────

-- name: GetLoginRateLimit :one
SELECT * FROM login_rate_limits
WHERE scope_type = $1 AND scope_key = $2;

-- name: UpsertLoginRateLimit :one
INSERT INTO login_rate_limits (scope_type, scope_key, tokens, updated_at)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (scope_type, scope_key) DO UPDATE
    SET tokens = $3, updated_at = NOW()
RETURNING *;

-- name: DeleteLoginRateLimit :exec
DELETE FROM login_rate_limits WHERE scope_type = $1 AND scope_key = $2;
