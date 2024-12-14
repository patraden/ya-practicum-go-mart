-- name: CreateUser :one
INSERT INTO users (id, username, password, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (username) DO UPDATE
SET id = users.id,
    username = users.username,
    password = users.password,
    created_at = users.created_at,
    updated_at = users.updated_at
RETURNING id, username, password, created_at, updated_at;

-- name: GetUser :one
SELECT id, username, password, created_at, updated_at
FROM users
WHERE username = $1;