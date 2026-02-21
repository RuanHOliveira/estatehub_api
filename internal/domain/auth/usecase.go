package auth

import (
	"context"

	errors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/security"
	repo "github.com/RuanHOliveira/estatehub_api/internal/infra/database/postgresql/sqlc/generated"
	"golang.org/x/crypto/bcrypt"
)

type TxManager interface {
	WithTx(ctx context.Context, fn func(q repo.Querier) error) error
}

type AuthUsecase interface {
	Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error)
	Login(ctx context.Context, input *LoginInput) (*LoginOutput, error)
}

type uc struct {
	txm TxManager
	j   security.JwtService
}

func NewAuthUsecase(txm TxManager, j security.JwtService) AuthUsecase {
	return &uc{txm: txm, j: j}
}

func (u *uc) Register(ctx context.Context, input *RegisterInput) (*RegisterOutput, error) {
	var output *RegisterOutput

	err := u.txm.WithTx(ctx, func(q repo.Querier) error {
		// Valida se já existe usuário com o email informado
		_, err := q.FindUserByEmail(ctx, input.Email)
		if err == nil {
			return errors.ErrEmailAlreadyUsed
		}

		// Faz hash da senha
		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 14)
		if err != nil {
			return err
		}

		// Cria usuário
		user, err := q.CreateUser(ctx, repo.CreateUserParams{Email: input.Email, Name: input.Name, PasswordHash: string(hash)})
		if err != nil {
			return err
		}

		// Cria token JWT
		accessToken, err := u.j.GenerateAccessToken(user.ID)
		if err != nil {
			return err
		}

		output = &RegisterOutput{User: UserOutput{ID: user.ID, Name: user.Name, Email: user.Email}, AccessToken: accessToken}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}

func (u *uc) Login(ctx context.Context, input *LoginInput) (*LoginOutput, error) {
	var output *LoginOutput

	err := u.txm.WithTx(ctx, func(q repo.Querier) error {
		// Busca usuário pelo email
		user, err := q.FindUserByEmail(ctx, input.Email)
		if err != nil {
			return errors.ErrUserNotFound
		}

		// Faz o hash da senha recebida e compara com hash da senha salva no DB
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
			return errors.ErrInvalidCredentials
		}

		// Cria token JWT
		accessToken, err := u.j.GenerateAccessToken(user.ID)
		if err != nil {
			return err
		}

		output = &LoginOutput{User: UserOutput{ID: user.ID, Name: user.Name, Email: user.Email}, AccessToken: accessToken}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return output, nil
}
