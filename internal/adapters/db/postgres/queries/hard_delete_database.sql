-- name: HardDeleteDatabase :exec
DELETE FROM databases WHERE id = $1;
