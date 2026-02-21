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

type CreatePropertyAdOutput struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Type         string    `json:"type"`
	PriceBrl     float64   `json:"price_brl"`
	ImagePath    *string   `json:"image_path"`
	ZipCode      string    `json:"zip_code"`
	Street       string    `json:"street"`
	Number       string    `json:"number"`
	Neighborhood string    `json:"neighborhood"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Complement   *string   `json:"complement"`
}

type PropertyAdItem struct {
	ID           uuid.UUID `json:"id"`
	UserID       uuid.UUID `json:"user_id"`
	Type         string    `json:"type"`
	PriceBrl     float64   `json:"price_brl"`
	ImagePath    *string   `json:"image_path"`
	ImageData    *string   `json:"image_data"`
	ZipCode      string    `json:"zip_code"`
	Street       string    `json:"street"`
	Number       string    `json:"number"`
	Neighborhood string    `json:"neighborhood"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	Complement   *string   `json:"complement"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
