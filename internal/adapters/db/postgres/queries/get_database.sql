-- name: GetDatabase :one
SELECT * FROM databases WHERE id = $1;
