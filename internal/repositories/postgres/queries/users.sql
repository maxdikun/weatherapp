-- name: SelectUserById :one
SELECT id, login, password
FROM users
WHERE id = $1;

-- name: SelectUserByLogin :one
SELECT id, login, password
FROM users
WHERE login = $1;

-- name: InsertUser :one
INSERT INTO users (id, login, password)
VALUES ($1, $2, $3)
RETURNING *;