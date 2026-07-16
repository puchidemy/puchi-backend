-- name: CreateAuditLog :one
INSERT INTO auth.audit_logs (user_id, action, resource, resource_id, ip_address, user_agent, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING *;

-- name: ListAuditLogsByUser :many
SELECT * FROM auth.audit_logs WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2;
