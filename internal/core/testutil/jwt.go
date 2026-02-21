package testutil

import (
	"github.com/RuanHOliveira/estatehub_api/internal/core/security"
	"github.com/google/uuid"
)

type MockJwtService struct {
	GenerateFn func(userID uuid.UUID) (string, error)
}

var _ security.JwtService = (*MockJwtService)(nil)

func (m *MockJwtService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	return m.GenerateFn(userID)
}

func (m *MockJwtService) ValidateAccessToken(_ string) (*uuid.UUID, error) {
	panic("ValidateAccessToken não é esperado em testes de usecase")
}
