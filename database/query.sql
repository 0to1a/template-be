-- name: FindUserByEmail :one
SELECT id, email, name, token, created_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: FindUserByToken :one
SELECT id, email, name, created_at
FROM users
WHERE token = $1::text AND deleted_at IS NULL;

-- name: UpdateUserToken :exec
UPDATE users SET token = $1 WHERE id = $2;
