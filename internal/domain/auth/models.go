package auth

import "github.com/google/uuid"

// RegisterRequest representa os dados para criar uma nova conta.
type RegisterRequest struct {
	Email    string `json:"email" example:"usuario@email.com"`
	Name     string `json:"name" example:"João Silva"`
	Password string `json:"password" example:"senha123"`
}

// RegisterResponse é retornado após registro bem-sucedido.
type RegisterResponse struct {
	User        UserOutput `json:"user"`
	AccessToken string     `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1NTBlODQwMCJ9.abc"`
}

type RegisterInput struct {
	Email    string
	Name     string
	Password string
}

type RegisterOutput struct {
	User        UserOutput
	AccessToken string
}

// LoginRequest representa as credenciais de autenticação.
type LoginRequest struct {
	Email    string `json:"email" example:"usuario@email.com"`
	Password string `json:"password" example:"senha123"`
}

// LoginResponse é retornado após autenticação bem-sucedida.
type LoginResponse struct {
	User        UserOutput `json:"user"`
	AccessToken string     `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiI1NTBlODQwMCJ9.abc"`
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	User        UserOutput
	AccessToken string
}

// UserOutput representa os dados públicos do usuário.
type UserOutput struct {
	ID    uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email string    `json:"email" example:"usuario@email.com"`
	Name  string    `json:"name" example:"João Silva"`
}
