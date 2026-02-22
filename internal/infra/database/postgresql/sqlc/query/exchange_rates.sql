-- name: CreateExchangeRate :one
INSERT INTO exchange_rates (user_id, target_currency, rate) VALUES ($1, $2, $3) RETURNING *;

-- name: ListAllExchangeRates :many
SELECT * FROM exchange_rates ORDER BY created_at DESC;

-- name: SoftDeleteAllExchangeRates :exec
UPDATE exchange_rates SET deleted_at = NOW(), updated_at = NOW() WHERE deleted_at IS NULL;

-- name: GetActiveExchangeRate :one
SELECT id, user_id, target_currency, rate, created_at, updated_at, deleted_at
FROM exchange_rates
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 1;