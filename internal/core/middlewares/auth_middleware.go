package middlewares

import (
	"context"
	"net/http"
	"strings"

	errors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/json"
	"github.com/RuanHOliveira/estatehub_api/internal/core/security"
)

type contextKey string

const UserIDContextKey contextKey = "user_id"

func AuthMiddleware(jwtService security.JwtService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				json.WriteError(w, http.StatusUnauthorized, errors.ErrMissingToken)
				return
			}

			const bearer = "Bearer "
			if !strings.HasPrefix(authHeader, bearer) {
				json.WriteError(w, http.StatusUnauthorized, errors.ErrInvalidToken)
				return
			}

			token := strings.TrimPrefix(authHeader, bearer)

			userID, err := jwtService.ValidateAccessToken(token)
			if err != nil {
				json.WriteError(w, http.StatusUnauthorized, errors.ErrInvalidToken)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDContextKey, *userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
