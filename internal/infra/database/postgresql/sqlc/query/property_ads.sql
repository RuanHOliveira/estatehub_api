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
