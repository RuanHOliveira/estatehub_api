package property_ads_test

import (
	"context"
	"errors"
	"testing"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/testutil"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/property_ads"
	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockTxManager struct {
	q     repo.Querier
	txErr error
}

func (m *mockTxManager) WithTx(ctx context.Context, fn func(q repo.Querier) error) error {
	if m.txErr != nil {
		return m.txErr
	}
	return fn(m.q)
}

type mockQuerier struct {
	createPropertyAdFn      func(ctx context.Context, arg repo.CreatePropertyAdParams) (repo.PropertyAd, error)
	listPropertyAdsFn       func(ctx context.Context) ([]repo.PropertyAd, error)
	getActiveExchangeRateFn func(ctx context.Context) (repo.ExchangeRate, error)
	softDeletePropertyAdFn  func(ctx context.Context, id uuid.UUID) (int64, error)
}

func (m *mockQuerier) CreatePropertyAd(ctx context.Context, arg repo.CreatePropertyAdParams) (repo.PropertyAd, error) {
	return m.createPropertyAdFn(ctx, arg)
}

func (m *mockQuerier) ListPropertyAds(ctx context.Context) ([]repo.PropertyAd, error) {
	return m.listPropertyAdsFn(ctx)
}

func (m *mockQuerier) GetActiveExchangeRate(ctx context.Context) (repo.ExchangeRate, error) {
	if m.getActiveExchangeRateFn != nil {
		return m.getActiveExchangeRateFn(ctx)
	}
	return repo.ExchangeRate{}, errors.New("sem cotação ativa")
}

func (m *mockQuerier) SoftDeletePropertyAd(ctx context.Context, id uuid.UUID) (int64, error) {
	return m.softDeletePropertyAdFn(ctx, id)
}

func (m *mockQuerier) CreateUser(_ context.Context, _ repo.CreateUserParams) (repo.User, error) {
	panic("CreateUser não é esperado em testes de property_ads")
}

func (m *mockQuerier) FindUserByEmail(_ context.Context, _ string) (repo.User, error) {
	panic("FindUserByEmail não é esperado em testes de property_ads")
}

func (m *mockQuerier) CreateExchangeRate(_ context.Context, _ repo.CreateExchangeRateParams) (repo.ExchangeRate, error) {
	panic("CreateExchangeRate não é esperado em testes de property_ads")
}

func (m *mockQuerier) ListAllExchangeRates(ctx context.Context) ([]repo.ExchangeRate, error) {
	panic("ListAllExchangeRates não é esperado em testes de property_ads")
}

func (m *mockQuerier) SoftDeleteAllExchangeRates(_ context.Context) error {
	panic("SoftDeleteAllExchangeRates não é esperado em testes de property_ads")
}

func validInput() *property_ads.CreatePropertyAdInput {
	return &property_ads.CreatePropertyAdInput{
		UserID:       testutil.FixedUserID,
		Type:         "SALE",
		PriceBrl:     500000.00,
		ImagePath:    "/uploads/standart_property_image.jpg",
		ZipCode:      "01310-100",
		Street:       "Av. Paulista",
		Number:       "1000",
		Neighborhood: "Bela Vista",
		City:         "São Paulo",
		State:        "SP",
	}
}

func fixedPropertyAd() repo.PropertyAd {
	var price pgtype.Numeric
	price.Scan("500000.00")
	imagePath := "/uploads/standart_property_image.jpg"
	return repo.PropertyAd{
		ID:           uuid.New(),
		UserID:       testutil.FixedUserID,
		Type:         "SALE",
		PriceBrl:     price,
		ImagePath:    &imagePath,
		ZipCode:      "01310-100",
		Street:       "Av. Paulista",
		Number:       "1000",
		Neighborhood: "Bela Vista",
		City:         "São Paulo",
		State:        "SP",
	}
}

func fixedExchangeRate() repo.ExchangeRate {
	var rate pgtype.Numeric
	rate.Scan("5.00")
	return repo.ExchangeRate{
		ID:             uuid.New(),
		UserID:         testutil.FixedUserID,
		TargetCurrency: "USD",
		Rate:           rate,
	}
}

func TestPropertyAdUsecase_CreatePropertyAd(t *testing.T) {
	tests := []struct {
		name        string
		inputFn     func() *property_ads.CreatePropertyAdInput
		querier     *mockQuerier
		txErr       error
		wantErr     error
		checkOutput func(t *testing.T, out *property_ads.CreatePropertyAdOutput)
	}{
		{
			name:    "sucesso com cotação ativa",
			inputFn: validInput,
			querier: &mockQuerier{
				createPropertyAdFn: func(_ context.Context, _ repo.CreatePropertyAdParams) (repo.PropertyAd, error) {
					return fixedPropertyAd(), nil
				},
				getActiveExchangeRateFn: func(_ context.Context) (repo.ExchangeRate, error) {
					return fixedExchangeRate(), nil
				},
			},
			checkOutput: func(t *testing.T, out *property_ads.CreatePropertyAdOutput) {
				if out == nil {
					t.Fatal("output não deveria ser nil")
				}
				if out.ID == uuid.Nil {
					t.Error("ID do anúncio não deveria ser nulo")
				}
				if out.ImagePath == nil {
					t.Error("image_path não deveria ser nil")
				}
				if out.PriceUsd == nil {
					t.Error("price_usd não deveria ser nil quando há cotação ativa")
				} else if *out.PriceUsd != 500000.00*5.00 {
					t.Errorf("price_usd: esperado %f, recebido %f", 500000.00*5.00, *out.PriceUsd)
				}
			},
		},
		{
			name:    "sucesso sem cotação ativa",
			inputFn: validInput,
			querier: &mockQuerier{
				createPropertyAdFn: func(_ context.Context, _ repo.CreatePropertyAdParams) (repo.PropertyAd, error) {
					return fixedPropertyAd(), nil
				},
			},
			checkOutput: func(t *testing.T, out *property_ads.CreatePropertyAdOutput) {
				if out == nil {
					t.Fatal("output não deveria ser nil")
				}
				if out.PriceUsd != nil {
					t.Errorf("price_usd deveria ser nil quando não há cotação ativa, recebeu %f", *out.PriceUsd)
				}
			},
		},
		{
			name: "tipo inválido",
			inputFn: func() *property_ads.CreatePropertyAdInput {
				i := validInput()
				i.Type = "LEASE"
				return i
			},
			querier: &mockQuerier{},
			wantErr: coreerrors.ErrInvalidAdType,
		},
		{
			name: "preço zero",
			inputFn: func() *property_ads.CreatePropertyAdInput {
				i := validInput()
				i.PriceBrl = 0
				return i
			},
			querier: &mockQuerier{},
			wantErr: coreerrors.ErrInvalidPrice,
		},
		{
			name: "preço negativo",
			inputFn: func() *property_ads.CreatePropertyAdInput {
				i := validInput()
				i.PriceBrl = -1
				return i
			},
			querier: &mockQuerier{},
			wantErr: coreerrors.ErrInvalidPrice,
		},
		{
			name: "campo de endereço vazio",
			inputFn: func() *property_ads.CreatePropertyAdInput {
				i := validInput()
				i.City = ""
				return i
			},
			querier: &mockQuerier{},
			wantErr: coreerrors.ErrMissingAddressField,
		},
		{
			name:    "erro no repositório",
			inputFn: validInput,
			querier: &mockQuerier{
				createPropertyAdFn: func(_ context.Context, _ repo.CreatePropertyAdParams) (repo.PropertyAd, error) {
					return repo.PropertyAd{}, errors.New("conexão com banco recusada")
				},
			},
			wantErr: errors.New("conexão com banco recusada"),
		},
		{
			name:    "erro ao iniciar transação",
			inputFn: validInput,
			querier: &mockQuerier{},
			txErr:   errors.New("pool de conexões esgotado"),
			wantErr: errors.New("pool de conexões esgotado"),
		},
		{
			name: "sucesso sem imagem",
			inputFn: func() *property_ads.CreatePropertyAdInput {
				i := validInput()
				i.ImagePath = ""
				return i
			},
			querier: &mockQuerier{
				createPropertyAdFn: func(_ context.Context, _ repo.CreatePropertyAdParams) (repo.PropertyAd, error) {
					ad := fixedPropertyAd()
					ad.ImagePath = nil
					return ad, nil
				},
			},
			checkOutput: func(t *testing.T, out *property_ads.CreatePropertyAdOutput) {
				if out == nil {
					t.Fatal("output não deveria ser nil")
				}
				if out.ImagePath != nil {
					t.Errorf("image_path deveria ser nil quando sem imagem, recebeu %q", *out.ImagePath)
				}
			},
		},
		{
			name: "sucesso com imagem",
			inputFn: func() *property_ads.CreatePropertyAdInput {
				i := validInput()
				i.ImagePath = "/uploads/property_ads/abc.jpg"
				return i
			},
			querier: &mockQuerier{
				createPropertyAdFn: func(_ context.Context, arg repo.CreatePropertyAdParams) (repo.PropertyAd, error) {
					ad := fixedPropertyAd()
					ad.ImagePath = arg.ImagePath
					return ad, nil
				},
			},
			checkOutput: func(t *testing.T, out *property_ads.CreatePropertyAdOutput) {
				if out.ImagePath == nil {
					t.Error("image_path não deveria ser nil quando imagem foi enviada")
				}
				if *out.ImagePath != "/uploads/property_ads/abc.jpg" {
					t.Errorf("image_path: esperado %q, recebido %q", "/uploads/property_ads/abc.jpg", *out.ImagePath)
				}
			},
		},
		{
			name: "sucesso com complemento",
			inputFn: func() *property_ads.CreatePropertyAdInput {
				i := validInput()
				i.Complement = "Apto 42"
				return i
			},
			querier: &mockQuerier{
				createPropertyAdFn: func(_ context.Context, arg repo.CreatePropertyAdParams) (repo.PropertyAd, error) {
					ad := fixedPropertyAd()
					ad.Complement = arg.Complement
					return ad, nil
				},
			},
			checkOutput: func(t *testing.T, out *property_ads.CreatePropertyAdOutput) {
				if out.Complement == nil {
					t.Error("complement não deveria ser nil quando informado")
				}
				if *out.Complement != "Apto 42" {
					t.Errorf("complement: esperado %q, recebido %q", "Apto 42", *out.Complement)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			txm := &mockTxManager{q: tc.querier, txErr: tc.txErr}
			uc := property_ads.NewPropertyAdUsecase(txm)

			out, err := uc.CreatePropertyAd(context.Background(), tc.inputFn())

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("esperava erro %q, mas não recebeu nenhum", tc.wantErr)
				}
				if err.Error() != tc.wantErr.Error() {
					t.Errorf("erro: esperado %q, recebido %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}

			if tc.checkOutput != nil {
				tc.checkOutput(t, out)
			}
		})
	}
}

func TestPropertyAdUsecase_ListPropertyAds(t *testing.T) {
	tests := []struct {
		name        string
		querier     *mockQuerier
		txErr       error
		wantErr     error
		checkOutput func(t *testing.T, out []property_ads.PropertyAdItem)
	}{
		{
			name: "sucesso com cotação ativa",
			querier: &mockQuerier{
				listPropertyAdsFn: func(_ context.Context) ([]repo.PropertyAd, error) {
					return []repo.PropertyAd{fixedPropertyAd()}, nil
				},
				getActiveExchangeRateFn: func(_ context.Context) (repo.ExchangeRate, error) {
					return fixedExchangeRate(), nil
				},
			},
			checkOutput: func(t *testing.T, out []property_ads.PropertyAdItem) {
				if len(out) != 1 {
					t.Fatalf("esperava 1 item, recebeu %d", len(out))
				}
				if out[0].ID == uuid.Nil {
					t.Error("ID do anúncio não deveria ser nulo")
				}
				if out[0].PriceBrl != 500000.0 {
					t.Errorf("price_brl: esperado 500000.0, recebido %f", out[0].PriceBrl)
				}
				if out[0].ImagePath == nil {
					t.Error("image_path não deveria ser nil")
				}
				if out[0].PriceUsd == nil {
					t.Error("price_usd não deveria ser nil quando há cotação ativa")
				} else if *out[0].PriceUsd != 500000.0*5.00 {
					t.Errorf("price_usd: esperado %f, recebido %f", 500000.0*5.00, *out[0].PriceUsd)
				}
			},
		},
		{
			name: "sucesso sem cotação ativa",
			querier: &mockQuerier{
				listPropertyAdsFn: func(_ context.Context) ([]repo.PropertyAd, error) {
					return []repo.PropertyAd{fixedPropertyAd()}, nil
				},
			},
			checkOutput: func(t *testing.T, out []property_ads.PropertyAdItem) {
				if len(out) != 1 {
					t.Fatalf("esperava 1 item, recebeu %d", len(out))
				}
				if out[0].PriceUsd != nil {
					t.Errorf("price_usd deveria ser nil quando não há cotação ativa, recebeu %f", *out[0].PriceUsd)
				}
			},
		},
		{
			name: "lista vazia",
			querier: &mockQuerier{
				listPropertyAdsFn: func(_ context.Context) ([]repo.PropertyAd, error) {
					return []repo.PropertyAd{}, nil
				},
			},
			checkOutput: func(t *testing.T, out []property_ads.PropertyAdItem) {
				if len(out) != 0 {
					t.Errorf("esperava lista vazia, recebeu %d itens", len(out))
				}
			},
		},
		{
			name: "erro no repositório",
			querier: &mockQuerier{
				listPropertyAdsFn: func(_ context.Context) ([]repo.PropertyAd, error) {
					return nil, errors.New("falha na consulta ao banco")
				},
			},
			wantErr: errors.New("falha na consulta ao banco"),
		},
		{
			name:    "erro ao iniciar transação",
			querier: &mockQuerier{},
			txErr:   errors.New("pool de conexões esgotado"),
			wantErr: errors.New("pool de conexões esgotado"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			txm := &mockTxManager{q: tc.querier, txErr: tc.txErr}
			uc := property_ads.NewPropertyAdUsecase(txm)

			out, err := uc.ListPropertyAds(context.Background())

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("esperava erro %q, mas não recebeu nenhum", tc.wantErr)
				}
				if err.Error() != tc.wantErr.Error() {
					t.Errorf("erro: esperado %q, recebido %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}

			if tc.checkOutput != nil {
				tc.checkOutput(t, out)
			}
		})
	}
}

func TestPropertyAdUsecase_DeletePropertyAd(t *testing.T) {
	fixedID := uuid.New()

	tests := []struct {
		name    string
		id      uuid.UUID
		querier *mockQuerier
		txErr   error
		wantErr error
	}{
		{
			name: "sucesso",
			id:   fixedID,
			querier: &mockQuerier{
				softDeletePropertyAdFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
					return 1, nil
				},
			},
		},
		{
			name: "não encontrado ou já deletado",
			id:   fixedID,
			querier: &mockQuerier{
				softDeletePropertyAdFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
					return 0, nil
				},
			},
			wantErr: coreerrors.ErrPropertyAdNotFound,
		},
		{
			name: "erro no repositório",
			id:   fixedID,
			querier: &mockQuerier{
				softDeletePropertyAdFn: func(_ context.Context, _ uuid.UUID) (int64, error) {
					return 0, errors.New("falha ao executar update no banco")
				},
			},
			wantErr: errors.New("falha ao executar update no banco"),
		},
		{
			name:    "erro ao iniciar transação",
			id:      fixedID,
			querier: &mockQuerier{},
			txErr:   errors.New("pool de conexões esgotado"),
			wantErr: errors.New("pool de conexões esgotado"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			txm := &mockTxManager{q: tc.querier, txErr: tc.txErr}
			uc := property_ads.NewPropertyAdUsecase(txm)

			err := uc.DeletePropertyAd(context.Background(), tc.id)

			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("esperava erro %q, mas não recebeu nenhum", tc.wantErr)
				}
				if err.Error() != tc.wantErr.Error() {
					t.Errorf("erro: esperado %q, recebido %q", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("erro inesperado: %v", err)
			}
		})
	}
}
