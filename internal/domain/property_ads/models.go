package property_ads

import (
	"time"

	"github.com/google/uuid"
)

type CreatePropertyAdRequest struct {
	Type         string
	PriceBrlStr  string
	ZipCode      string
	Street       string
	Number       string
	Neighborhood string
	City         string
	State        string
	Complement   string
}

type CreatePropertyAdInput struct {
	UserID       uuid.UUID
	Type         string
	PriceBrl     float64
	ImagePath    string
	ZipCode      string
	Street       string
	Number       string
	Neighborhood string
	City         string
	State        string
	Complement   string
}

// CreatePropertyAdOutput é retornado após criação bem-sucedida de um anúncio.
type CreatePropertyAdOutput struct {
	ID           uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID       uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Type         string    `json:"type" example:"SALE"`
	PriceBrl     float64   `json:"price_brl" example:"450000.00"`
	PriceUsd     *float64  `json:"price_usd" example:"81600.00"`
	ImagePath    *string   `json:"image_path" example:"/uploads/property_ads/abc123.jpg"`
	ImageData    *string   `json:"image_data" example:"data:image/jpeg;base64,/9j/4AAQ..."`
	ZipCode      string    `json:"zip_code" example:"01310100"`
	Street       string    `json:"street" example:"Avenida Paulista"`
	Number       string    `json:"number" example:"1000"`
	Neighborhood string    `json:"neighborhood" example:"Bela Vista"`
	City         string    `json:"city" example:"São Paulo"`
	State        string    `json:"state" example:"SP"`
	Complement   *string   `json:"complement" example:"Apto 42"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PropertyAdItem representa um anúncio imobiliário na listagem.
type PropertyAdItem struct {
	ID           uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID       uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Type         string    `json:"type" example:"RENT"`
	PriceBrl     float64   `json:"price_brl" example:"2500.00"`
	PriceUsd     *float64  `json:"price_usd" example:"454.50"`
	ImagePath    *string   `json:"image_path" example:"/uploads/property_ads/abc123.jpg"`
	ImageData    *string   `json:"image_data" example:"data:image/jpeg;base64,/9j/4AAQ..."`
	ZipCode      string    `json:"zip_code" example:"01310100"`
	Street       string    `json:"street" example:"Avenida Paulista"`
	Number       string    `json:"number" example:"1000"`
	Neighborhood string    `json:"neighborhood" example:"Bela Vista"`
	City         string    `json:"city" example:"São Paulo"`
	State        string    `json:"state" example:"SP"`
	Complement   *string   `json:"complement" example:"Apto 42"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
