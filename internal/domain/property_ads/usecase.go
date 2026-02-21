package property_ads

import (
	"context"
	"fmt"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"github.com/jackc/pgx/v5/pgtype"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(q repo.Querier) error) error
}

type PropertyAdUsecase interface {
	CreatePropertyAd(ctx context.Context, input *CreatePropertyAdInput) (*CreatePropertyAdOutput, error)
}

type uc struct {
	txm TxManager
}

func NewPropertyAdUsecase(txm TxManager) PropertyAdUsecase {
	return &uc{txm: txm}
}

func (u *uc) CreatePropertyAd(ctx context.Context, input *CreatePropertyAdInput) (*CreatePropertyAdOutput, error) {
	if input.Type != "SALE" && input.Type != "RENT" {
		return nil, coreerrors.ErrInvalidAdType
	}

	if input.PriceBrl <= 0 {
		return nil, coreerrors.ErrInvalidPrice
	}

	if input.ZipCode == "" || input.Street == "" || input.Number == "" ||
		input.Neighborhood == "" || input.City == "" || input.State == "" {
		return nil, coreerrors.ErrMissingAddressField
	}

	var output *CreatePropertyAdOutput

	err := u.txm.WithTx(ctx, func(q repo.Querier) error {
		var price pgtype.Numeric
		if err := price.Scan(fmt.Sprintf("%.2f", input.PriceBrl)); err != nil {
			return err
		}

		var imagePath *string
		if input.ImagePath != "" {
			imagePath = &input.ImagePath
		}

		var complement *string
		if input.Complement != "" {
			complement = &input.Complement
		}

		ad, err := q.CreatePropertyAd(ctx, repo.CreatePropertyAdParams{
			UserID:       input.UserID,
			Type:         input.Type,
			PriceBrl:     price,
			ImagePath:    imagePath,
			ZipCode:      input.ZipCode,
			Street:       input.Street,
			Number:       input.Number,
			Neighborhood: input.Neighborhood,
			City:         input.City,
			State:        input.State,
			Complement:   complement,
		})
		if err != nil {
			return err
		}

		priceFloat, _ := ad.PriceBrl.Float64Value()

		output = &CreatePropertyAdOutput{
			ID:           ad.ID,
			UserID:       ad.UserID,
			Type:         ad.Type,
			PriceBrl:     priceFloat.Float64,
			ImagePath:    ad.ImagePath,
			ZipCode:      ad.ZipCode,
			Street:       ad.Street,
			Number:       ad.Number,
			Neighborhood: ad.Neighborhood,
			City:         ad.City,
			State:        ad.State,
			Complement:   ad.Complement,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
