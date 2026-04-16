-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, username, email, password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2,
    $3
)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;



-- name: UpdateUser :one
UPDATE users SET username = $2, email = $3, password = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUsers :many
SELECT * FROM users
ORDER BY created_at ASC;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1;