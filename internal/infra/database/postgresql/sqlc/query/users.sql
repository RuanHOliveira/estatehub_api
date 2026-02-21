-- name: FindUserByEmail :one
SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL;

-- name: CreateUser :one
INSERT INTO
    users (
        email,
        name,
        password_hash
    )
VALUES ($1, $2, $3) RETURNING *;