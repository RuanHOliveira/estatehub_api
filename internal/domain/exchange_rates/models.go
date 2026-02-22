package exchange_rates

import (
	"time"

	"github.com/google/uuid"
)

type CreateExchangeRateRequest struct {
	TargetCurrency string `json:"target_currency"`
	RateStr        string `json:"rate"`
}

type CreateExchangeRateInput struct {
	UserID         uuid.UUID
	TargetCurrency string
	Rate           float64
}

type CreateExchangeRateOutput struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	TargetCurrency string    `json:"target_currency"`
	Rate           float64   `json:"rate"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ExchangeRateItem struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	TargetCurrency string    `json:"target_currency"`
	Rate           float64   `json:"rate"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
}
