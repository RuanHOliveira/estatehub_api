package auth

import (
	"context"

	"github.com/RuanHOliveira/estatehub_api/internal/core/config"
	errors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/security"
	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecase interface {
	Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input *LoginInput) (*LoginOutput, error)
}

type uc struct {
	r       *repo.Queries
	db      *pgx.Conn
	authCfg *config.AuthConfig
	j       security.JwtService
}

func NewAuthUsecase(r *repo.Queries, db *pgx.Conn, authCfg *config.AuthConfig, j security.JwtService) AuthUsecase {
	return &uc{r: r, db: db, authCfg: authCfg, j: j}
}

func (u *uc) Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)
	rtx := u.r.WithTx(tx)

	// Valida se já existe usuário com o email informado
	_, err = rtx.FindUserByEmail(ctx, input.Email)
	if err == nil {
		return nil, errors.ErrEmailAlreadyUsed
	}

	// Faz hash da senha
	bytes, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
	if err != nil {
		return nil, err
	}

	// Cria usuário
	user, err := rtx.CreateUser(ctx, repo.CreateUserParams{Email: input.Email, Name: input.Name, PasswordHash: string(bytes)})
	if err != nil {
		return nil, err
	}

	// Cria token JWT
	accessToken, err := u.j.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Mantem alterações no banco caso não ocorra erro durante o processo
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &RegisterOutput{User: UserOutput{ID: user.ID, Name: user.Name, Email: user.Email}, AccessToken: accessToken}, nil
}

func (u *uc) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)
	rtx := u.r.WithTx(tx)

	// Busca usuário pelo email
	user, err := rtx.FindUserByEmail(ctx, input.Email)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	// Faz o hash da senha recebida e compara com hash da senha salva no DB
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return nil, errors.ErrInvalidCredentials
	}

	// Cria token JWT
	accessToken, err := u.j.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Mantem alterações no banco caso não ocorra erro durante o processo
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &LoginOutput{User: UserOutput{ID: user.ID, Name: user.Name, Email: user.Email}, AccessToken: accessToken}, nil
}
