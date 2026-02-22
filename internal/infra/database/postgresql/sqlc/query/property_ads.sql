-- name: CreatePropertyAd :one
INSERT INTO 
    property_ads (
        user_id, 
        type, 
        price_brl, 
        image_path,
        zip_code, 
        street, 
        number, 
        neighborhood, 
        city, 
        state, 
        complement
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: ListPropertyAds :many
SELECT id, user_id, type, price_brl, image_path, zip_code, street, number, neighborhood, city, state, complement, created_at, updated_at, deleted_at
FROM property_ads
WHERE deleted_at IS NULL
ORDER BY created_at DESC;

-- name: SoftDeletePropertyAd :execrows
UPDATE property_ads
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;
