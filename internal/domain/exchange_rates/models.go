package exchange_rates

import (
	"time"

	"github.com/google/uuid"
)

// CreateExchangeRateRequest representa os dados para criar uma nova cotação.
type CreateExchangeRateRequest struct {
	TargetCurrency string `json:"target_currency" example:"USD"`
	RateStr        string `json:"rate" example:"0.181818"`
}

type CreateExchangeRateInput struct {
	UserID         uuid.UUID
	TargetCurrency string
	Rate           float64
}

// CreateExchangeRateOutput é retornado após criação bem-sucedida de uma cotação.
type CreateExchangeRateOutput struct {
	ID             uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID         uuid.UUID `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	TargetCurrency string    `json:"target_currency" example:"USD"`
	Rate           float64   `json:"rate" example:"0.181818"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ExchangeRateItem representa uma cotação no histórico (ativa ou inativa).
type ExchangeRateItem struct {
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID         uuid.UUID  `json:"user_id" example:"550e8400-e29b-41d4-a716-446655440001"`
	TargetCurrency string     `json:"target_currency" example:"USD"`
	Rate           float64    `json:"rate" example:"0.181818"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	DeletedAt      *time.Time `json:"deleted_at"`
}
