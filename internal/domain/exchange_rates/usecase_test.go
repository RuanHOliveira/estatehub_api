package exchange_rates_test

import (
	"context"
	"errors"
	"testing"
	"time"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/testutil"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/exchange_rates"
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
	createExchangeRateFn   func(ctx context.Context, arg repo.CreateExchangeRateParams) (repo.ExchangeRate, error)
	listAllExchangeRatesFn func(ctx context.Context) ([]repo.ExchangeRate, error)
	deleteExchangeRatesFn  func(ctx context.Context) error
}

func (m *mockQuerier) CreateExchangeRate(ctx context.Context, arg repo.CreateExchangeRateParams) (repo.ExchangeRate, error) {
	return m.createExchangeRateFn(ctx, arg)
}

func (m *mockQuerier) ListAllExchangeRates(ctx context.Context) ([]repo.ExchangeRate, error) {
	return m.listAllExchangeRatesFn(ctx)
}

func (m *mockQuerier) DeleteExchangeRates(ctx context.Context) error {
	if m.deleteExchangeRatesFn != nil {
		return m.deleteExchangeRatesFn(ctx)
	}
	return nil
}

func (m *mockQuerier) CreatePropertyAd(ctx context.Context, arg repo.CreatePropertyAdParams) (repo.PropertyAd, error) {
	panic("CreatePropertyAd não é esperado em testes de exchange_rates")
}

func (m *mockQuerier) ListPropertyAds(ctx context.Context) ([]repo.PropertyAd, error) {
	panic("ListPropertyAds não é esperado em testes de exchange_rates")
}

func (m *mockQuerier) CreateUser(_ context.Context, _ repo.CreateUserParams) (repo.User, error) {
	panic("CreateUser não é esperado em testes de exchange_rates")
}

func (m *mockQuerier) FindUserByEmail(_ context.Context, _ string) (repo.User, error) {
	panic("FindUserByEmail não é esperado em testes de exchange_rates")
}

func validInput() *exchange_rates.CreateExchangeRateInput {
	return &exchange_rates.CreateExchangeRateInput{
		UserID:         testutil.FixedUserID,
		TargetCurrency: "USD",
		Rate:           2.50,
	}
}

func fixedExchangeRate() repo.ExchangeRate {
	var rate pgtype.Numeric
	rate.Scan("2.50")
	return repo.ExchangeRate{
		ID:             uuid.New(),
		UserID:         testutil.FixedUserID,
		TargetCurrency: "USD",
		Rate:           rate,
	}
}

func TestExchangeRateUsecase_CreateExchangeRate(t *testing.T) {
	tests := []struct {
		name        string
		inputFn     func() *exchange_rates.CreateExchangeRateInput
		querier     *mockQuerier
		txErr       error
		wantErr     error
		checkOutput func(t *testing.T, out *exchange_rates.CreateExchangeRateOutput)
	}{
		{
			name:    "sucesso",
			inputFn: validInput,
			querier: &mockQuerier{
				createExchangeRateFn: func(_ context.Context, _ repo.CreateExchangeRateParams) (repo.ExchangeRate, error) {
					return fixedExchangeRate(), nil
				},
			},
			checkOutput: func(t *testing.T, out *exchange_rates.CreateExchangeRateOutput) {
				if out == nil {
					t.Fatal("output não deveria ser nil")
				}
				if out.ID == uuid.Nil {
					t.Error("ID da cotação não deveria ser nulo")
				}
				if out.TargetCurrency != "USD" {
					t.Errorf("target_currency: esperado %q, recebido %q", "USD", out.TargetCurrency)
				}
			},
		},
		{
			name: "rate zero",
			inputFn: func() *exchange_rates.CreateExchangeRateInput {
				i := validInput()
				i.Rate = 0
				return i
			},
			querier: &mockQuerier{},
			wantErr: coreerrors.ErrInvalidRate,
		},
		{
			name: "rate negativa",
			inputFn: func() *exchange_rates.CreateExchangeRateInput {
				i := validInput()
				i.Rate = -1.5
				return i
			},
			querier: &mockQuerier{},
			wantErr: coreerrors.ErrInvalidRate,
		},
		{
			name:    "erro no repositório ao criar",
			inputFn: validInput,
			querier: &mockQuerier{
				createExchangeRateFn: func(_ context.Context, _ repo.CreateExchangeRateParams) (repo.ExchangeRate, error) {
					return repo.ExchangeRate{}, errors.New("conexão com banco recusada")
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			txm := &mockTxManager{q: tc.querier, txErr: tc.txErr}
			uc := exchange_rates.NewExchangeRateUsecase(txm)

			out, err := uc.CreateExchangeRate(context.Background(), tc.inputFn())

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

func TestExchangeRateUsecase_ListAllExchangeRates(t *testing.T) {
	tests := []struct {
		name        string
		querier     *mockQuerier
		txErr       error
		wantErr     error
		checkOutput func(t *testing.T, out []exchange_rates.ExchangeRateItem)
	}{
		{
			name: "sucesso incluindo registros deletados",
			querier: &mockQuerier{
				listAllExchangeRatesFn: func(_ context.Context) ([]repo.ExchangeRate, error) {
					active := fixedExchangeRate()
					deletedTime := time.Now()
					deleted := fixedExchangeRate()
					deleted.ID = uuid.New()
					deleted.DeletedAt = pgtype.Timestamptz{Valid: true, Time: deletedTime}

					return []repo.ExchangeRate{active, deleted}, nil
				},
			},
			checkOutput: func(t *testing.T, out []exchange_rates.ExchangeRateItem) {
				if len(out) != 2 {
					t.Fatalf("esperava 2 itens (incluindo deletado), recebeu %d", len(out))
				}
				if out[0].DeletedAt != nil {
					t.Errorf("primeiro item não deveria ter deleted_at, recebeu %v", *out[0].DeletedAt)
				}
				if out[1].DeletedAt == nil {
					t.Error("segundo item deveria ter deleted_at preenchido")
				}
			},
		},
		{
			name: "lista vazia",
			querier: &mockQuerier{
				listAllExchangeRatesFn: func(_ context.Context) ([]repo.ExchangeRate, error) {
					return []repo.ExchangeRate{}, nil
				},
			},
			checkOutput: func(t *testing.T, out []exchange_rates.ExchangeRateItem) {
				if len(out) != 0 {
					t.Errorf("esperava lista vazia, recebeu %d itens", len(out))
				}
			},
		},
		{
			name: "erro no repositório",
			querier: &mockQuerier{
				listAllExchangeRatesFn: func(_ context.Context) ([]repo.ExchangeRate, error) {
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
			uc := exchange_rates.NewExchangeRateUsecase(txm)

			out, err := uc.ListAllExchangeRates(context.Background())

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
