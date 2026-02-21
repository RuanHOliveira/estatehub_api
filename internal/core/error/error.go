package errors

import "errors"

var (
	// Global
	ErrUnknown        = errors.New("ErrUnknown")        // Erro desconhecido
	ErrMissingToken   = errors.New("ErrMissingToken")   // Access Token faltando
	ErrInvalidToken   = errors.New("ErrInvalidToken")   // Access Token inválido
	ErrInvalidRequest = errors.New("ErrInvalidRequest") // Requisição inválida

	// Auth
	ErrEmailAlreadyUsed   = errors.New("ErrEmailAlreadyUsed")   // Email jé em uso
	ErrUserNotFound       = errors.New("ErrUserNotFound")       // Usuário não encontrado
	ErrInvalidCredentials = errors.New("ErrInvalidCredentials") // Credenciais inválidas
)
