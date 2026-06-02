-- name: CreateInvitation :one
INSERT INTO invitations (email, role, token_hash, invited_by, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetInvitationByID :one
SELECT * FROM invitations WHERE id = $1;

-- name: GetInvitationByTokenHash :one
SELECT * FROM invitations WHERE token_hash = $1;

-- name: ListPendingInvitations :many
SELECT * FROM invitations
WHERE accepted_at IS NULL AND expires_at > NOW()
ORDER BY created_at DESC;

-- name: AcceptInvitation :one
UPDATE invitations SET accepted_at = NOW() WHERE id = $1 RETURNING *;

-- name: DeleteInvitation :exec
DELETE FROM invitations WHERE id = $1;

-- name: DeleteExpiredInvitations :exec
DELETE FROM invitations WHERE expires_at < NOW() AND accepted_at IS NULL;
