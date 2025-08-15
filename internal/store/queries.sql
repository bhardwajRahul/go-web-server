-- name: GetUser :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users 
WHERE is_active = true 
ORDER BY created_at DESC;

-- name: ListAllUsers :many
SELECT * FROM users ORDER BY created_at DESC;

-- name: CreateUser :one
INSERT INTO users (email, name, bio, avatar_url, password_hash) 
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateUser :one
UPDATE users 
SET name = $1, bio = $2, avatar_url = $3, updated_at = CURRENT_TIMESTAMP
WHERE id = $4
RETURNING *;

-- name: DeactivateUser :exec
UPDATE users 
SET is_active = false, updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE is_active = true;