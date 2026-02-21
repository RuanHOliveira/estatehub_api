package security

import (
	"time"

	"github.com/RuanHOliveira/estatehub_api/internal/core/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtService interface {
	GenerateAccessToken(userID uuid.UUID) (string, error)
	ValidateAccessToken(tokenString string) (*uuid.UUID, error)
}

type jwtService struct {
	authConfig *config.AuthConfig
}

func NewJwtService(authConfig *config.AuthConfig) JwtService {
	return &jwtService{authConfig: authConfig}
}

func (j *jwtService) GenerateAccessToken(userID uuid.UUID) (string, error) {
	userIDString := userID.String()

	claims := jwt.MapClaims{
		"sub": userIDString,
		"typ": "access",
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.authConfig.JwtSecret))
}

func (j *jwtService) ValidateAccessToken(tokenString string) (*uuid.UUID, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(j.authConfig.JwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}

	// Valida tipo do token
	if claims["typ"] != "access" {
		return nil, jwt.ErrTokenInvalidClaims
	}

	sub, ok := claims["sub"].(string)
	if !ok {
		return nil, jwt.ErrTokenMalformed
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return nil, jwt.ErrTokenMalformed
	}

	return &userID, nil
}
