package auth

import "github.com/google/uuid"

type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	User        UserOutput `json:"user"`
	AccessToken string     `json:"access_token"`
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

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	User        UserOutput `json:"user"`
	AccessToken string     `json:"access_token"`
}

type LoginInput struct {
	Email    string
	Password string
}

type LoginOutput struct {
	User         UserOutput
	AccessToken string
}

type UserOutput struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Name  string    `json:"name"`
}
