package exchange_rates

import (
	"context"
	"fmt"
	"time"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(q repo.Querier) error) error
}

type ExchangeRateUsecase interface {
	CreateExchangeRate(ctx context.Context, input *CreateExchangeRateInput) (*CreateExchangeRateOutput, error)
	ListAllExchangeRates(ctx context.Context) ([]ExchangeRateItem, error)
}

type uc struct {
	txm TxManager
}

func NewExchangeRateUsecase(txm TxManager) ExchangeRateUsecase {
	return &uc{txm: txm}
}

func (u *uc) CreateExchangeRate(ctx context.Context, input *CreateExchangeRateInput) (*CreateExchangeRateOutput, error) {
	if input.Rate <= 0 {
		return nil, coreerrors.ErrInvalidRate
	}

	var output *CreateExchangeRateOutput

	err := u.txm.WithTx(ctx, func(q repo.Querier) error {
		var rate pgtype.Numeric
		if err := rate.Scan(fmt.Sprintf("%.6f", input.Rate)); err != nil {
			return err
		}

		// 'Inativa' todas as cotações cadastradas
		_ = q.SoftDeleteAllExchangeRates(ctx)

		er, err := q.CreateExchangeRate(ctx, repo.CreateExchangeRateParams{UserID: input.UserID, TargetCurrency: input.TargetCurrency, Rate: rate})
		if err != nil {
			return err
		}

		rateFloat, _ := er.Rate.Float64Value()

		output = &CreateExchangeRateOutput{ID: er.ID, UserID: er.UserID, TargetCurrency: er.TargetCurrency, Rate: rateFloat.Float64, CreatedAt: er.CreatedAt.Time, UpdatedAt: er.UpdatedAt.Time}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (u *uc) ListAllExchangeRates(ctx context.Context) ([]ExchangeRateItem, error) {
	var items []ExchangeRateItem

	err := u.txm.WithTx(ctx, func(q repo.Querier) error {
		rows, err := q.ListAllExchangeRates(ctx)
		if err != nil {
			return err
		}

		for _, row := range rows {
			rateFloat, _ := row.Rate.Float64Value()

			var deletedAt *time.Time
			if row.DeletedAt.Valid {
				t := row.DeletedAt.Time
				deletedAt = &t
			}

			items = append(items, ExchangeRateItem{
				ID:             row.ID,
				UserID:         row.UserID,
				TargetCurrency: row.TargetCurrency,
				Rate:           rateFloat.Float64,
				CreatedAt:      row.CreatedAt.Time,
				UpdatedAt:      row.UpdatedAt.Time,
				DeletedAt:      deletedAt,
			})
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return items, nil
}
