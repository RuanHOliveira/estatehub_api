package auth_test

import (
	"context"
	"errors"
	"testing"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/testutil"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/auth"
	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
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
	findUserByEmailFn func(ctx context.Context, email string) (repo.User, error)
	createUserFn      func(ctx context.Context, arg repo.CreateUserParams) (repo.User, error)
}

func (m *mockQuerier) FindUserByEmail(ctx context.Context, email string) (repo.User, error) {
	return m.findUserByEmailFn(ctx, email)
}

func (m *mockQuerier) CreateUser(ctx context.Context, arg repo.CreateUserParams) (repo.User, error) {
	return m.createUserFn(ctx, arg)
}

func (m *mockQuerier) CreatePropertyAd(_ context.Context, _ repo.CreatePropertyAdParams) (repo.PropertyAd, error) {
	panic("CreatePropertyAd não é esperado em testes de auth")
}

func (m *mockQuerier) ListPropertyAds(_ context.Context) ([]repo.PropertyAd, error) {
	panic("ListPropertyAds não é esperado em testes de auth")
}

func (m *mockQuerier) CreateExchangeRate(_ context.Context, _ repo.CreateExchangeRateParams) (repo.ExchangeRate, error) {
	panic("CreateExchangeRate não é esperado em testes de auth")
}

func (m *mockQuerier) ListAllExchangeRates(ctx context.Context) ([]repo.ExchangeRate, error) {
	panic("ListAllExchangeRates não é esperado em testes de auth")
}

func (m *mockQuerier) DeleteExchangeRates(_ context.Context) error {
	panic("DeleteExchangeRates não é esperado em testes de property_ads")
}

var loginTestPasswordHash string

func init() {
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		panic("falha ao gerar hash bcrypt para testes: " + err.Error())
	}
	loginTestPasswordHash = string(hash)
}

func TestAuthUsecase_Register(t *testing.T) {
	tests := []struct {
		name      string
		input     *auth.RegisterInput
		querier   *mockQuerier
		jwt       *testutil.MockJwtService
		txErr     error
		wantErr   error
		wantToken bool
	}{
		{
			name:  "sucesso",
			input: &auth.RegisterInput{Email: "new@example.com", Name: "New User", Password: "secret123"},
			querier: &mockQuerier{
				findUserByEmailFn: func(_ context.Context, _ string) (repo.User, error) {
					return repo.User{}, errors.New("não encontrado")
				},
				createUserFn: func(_ context.Context, arg repo.CreateUserParams) (repo.User, error) {
					return repo.User{ID: testutil.FixedUserID, Email: arg.Email, Name: arg.Name, PasswordHash: arg.PasswordHash}, nil
				},
			},
			jwt: &testutil.MockJwtService{
				GenerateFn: func(_ uuid.UUID) (string, error) { return "token-ok", nil },
			},
			wantToken: true,
		},
		{
			name:  "email já utilizado",
			input: &auth.RegisterInput{Email: "taken@example.com", Name: "Someone", Password: "secret123"},
			querier: &mockQuerier{
				findUserByEmailFn: func(_ context.Context, _ string) (repo.User, error) {
					return repo.User{ID: testutil.FixedUserID, Email: "taken@example.com"}, nil
				},
			},
			jwt:     &testutil.MockJwtService{},
			wantErr: coreerrors.ErrEmailAlreadyUsed,
		},
		{
			name:  "erro ao criar no repositório",
			input: &auth.RegisterInput{Email: "new@example.com", Name: "New User", Password: "secret123"},
			querier: &mockQuerier{
				findUserByEmailFn: func(_ context.Context, _ string) (repo.User, error) {
					return repo.User{}, errors.New("não encontrado")
				},
				createUserFn: func(_ context.Context, _ repo.CreateUserParams) (repo.User, error) {
					return repo.User{}, errors.New("conexão com banco recusada")
				},
			},
			jwt:     &testutil.MockJwtService{},
			wantErr: errors.New("conexão com banco recusada"),
		},
		{
			name:  "erro ao gerar token JWT",
			input: &auth.RegisterInput{Email: "new@example.com", Name: "New User", Password: "secret123"},
			querier: &mockQuerier{
				findUserByEmailFn: func(_ context.Context, _ string) (repo.User, error) {
					return repo.User{}, errors.New("não encontrado")
				},
				createUserFn: func(_ context.Context, arg repo.CreateUserParams) (repo.User, error) {
					return repo.User{ID: testutil.FixedUserID, Email: arg.Email, Name: arg.Name}, nil
				},
			},
			jwt: &testutil.MockJwtService{
				GenerateFn: func(_ uuid.UUID) (string, error) { return "", errors.New("falha ao assinar token JWT") },
			},
			wantErr: errors.New("falha ao assinar token JWT"),
		},
		{
			name:    "erro ao iniciar transação",
			input:   &auth.RegisterInput{Email: "new@example.com", Name: "New User", Password: "secret123"},
			querier: &mockQuerier{},
			jwt:     &testutil.MockJwtService{},
			txErr:   errors.New("pool de conexões esgotado"),
			wantErr: errors.New("pool de conexões esgotado"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			txm := &mockTxManager{q: tc.querier, txErr: tc.txErr}
			uc := auth.NewAuthUsecase(txm, tc.jwt)

			out, err := uc.Register(context.Background(), tc.input)

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

			if tc.wantToken {
				if out.AccessToken == "" {
					t.Error("access_token não deveria estar vazio")
				}
				if out.User.ID == uuid.Nil {
					t.Error("ID do usuário não deveria ser nulo")
				}
				if out.User.Email != tc.input.Email {
					t.Errorf("email do usuário: esperado %q, recebido %q", tc.input.Email, out.User.Email)
				}
			}
		})
	}
}

func TestAuthUsecase_Login(t *testing.T) {
	tests := []struct {
		name      string
		input     *auth.LoginInput
		querier   *mockQuerier
		jwt       *testutil.MockJwtService
		txErr     error
		wantErr   error
		wantToken bool
	}{
		{
			name:  "sucesso",
			input: &auth.LoginInput{Email: "user@example.com", Password: "correct-password"},
			querier: &mockQuerier{
				findUserByEmailFn: func(_ context.Context, _ string) (repo.User, error) {
					return repo.User{
						ID:           testutil.FixedUserID,
						Email:        "user@example.com",
						Name:         "Test User",
						PasswordHash: loginTestPasswordHash,
					}, nil
				},
			},
			jwt: &testutil.MockJwtService{
				GenerateFn: func(_ uuid.UUID) (string, error) { return "token-login-ok", nil },
			},
			wantToken: true,
		},
		{
			name:  "usuário não encontrado",
			input: &auth.LoginInput{Email: "ghost@example.com", Password: "whatever"},
			querier: &mockQuerier{
				findUserByEmailFn: func(_ context.Context, _ string) (repo.User, error) {
					return repo.User{}, errors.New("não encontrado")
				},
			},
			jwt:     &testutil.MockJwtService{},
			wantErr: coreerrors.ErrUserNotFound,
		},
		{
			name:  "senha inválida",
			input: &auth.LoginInput{Email: "user@example.com", Password: "wrong-password"},
			querier: &mockQuerier{
				findUserByEmailFn: func(_ context.Context, _ string) (repo.User, error) {
					return repo.User{
						ID:           testutil.FixedUserID,
						Email:        "user@example.com",
						PasswordHash: loginTestPasswordHash,
					}, nil
				},
			},
			jwt:     &testutil.MockJwtService{},
			wantErr: coreerrors.ErrInvalidCredentials,
		},
		{
			name:  "erro ao gerar token JWT",
			input: &auth.LoginInput{Email: "user@example.com", Password: "correct-password"},
			querier: &mockQuerier{
				findUserByEmailFn: func(_ context.Context, _ string) (repo.User, error) {
					return repo.User{ID: testutil.FixedUserID, PasswordHash: loginTestPasswordHash}, nil
				},
			},
			jwt: &testutil.MockJwtService{
				GenerateFn: func(_ uuid.UUID) (string, error) { return "", errors.New("falha ao assinar token JWT") },
			},
			wantErr: errors.New("falha ao assinar token JWT"),
		},
		{
			name:    "erro ao iniciar transação",
			input:   &auth.LoginInput{Email: "user@example.com", Password: "correct-password"},
			querier: &mockQuerier{},
			jwt:     &testutil.MockJwtService{},
			txErr:   errors.New("pool de conexões esgotado"),
			wantErr: errors.New("pool de conexões esgotado"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			txm := &mockTxManager{q: tc.querier, txErr: tc.txErr}
			uc := auth.NewAuthUsecase(txm, tc.jwt)

			out, err := uc.Login(context.Background(), tc.input)

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

			if tc.wantToken {
				if out.AccessToken == "" {
					t.Error("access_token não deveria estar vazio")
				}
				if out.User.ID == uuid.Nil {
					t.Error("ID do usuário não deveria ser nulo")
				}
			}
		})
	}
}
