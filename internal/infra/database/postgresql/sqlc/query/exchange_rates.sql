-- name: CreateExchangeRate :one
INSERT INTO exchange_rates (user_id, target_currency, rate) VALUES ($1, $2, $3) RETURNING *;

-- name: ListAllExchangeRates :many
SELECT * FROM exchange_rates ORDER BY created_at DESC;

-- name: DeleteExchangeRates :exec
UPDATE exchange_rates SET deleted_at = NOW(), updated_at = NOW() WHERE deleted_at IS NULL;