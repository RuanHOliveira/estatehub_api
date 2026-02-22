package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/testutil"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/auth"
	"github.com/google/uuid"
)

type mockAuthUsecase struct {
	registerFn func(ctx context.Context, input *auth.RegisterInput) (*auth.RegisterOutput, error)
	loginFn    func(ctx context.Context, input *auth.LoginInput) (*auth.LoginOutput, error)
}

func (m *mockAuthUsecase) Register(ctx context.Context, input *auth.RegisterInput) (*auth.RegisterOutput, error) {
	return m.registerFn(ctx, input)
}

func (m *mockAuthUsecase) Login(ctx context.Context, input *auth.LoginInput) (*auth.LoginOutput, error) {
	return m.loginFn(ctx, input)
}

func TestRegisterHandler(t *testing.T) {
	successOutput := &auth.RegisterOutput{
		User:        auth.UserOutput{ID: testutil.FixedUserID, Email: "user@example.com", Name: "Test User"},
		AccessToken: "token-register-ok",
	}

	tests := []struct {
		name          string
		body          any
		rawBody       string
		registerFn    func(context.Context, *auth.RegisterInput) (*auth.RegisterOutput, error)
		wantStatus    int
		wantErrorCode string
		wantToken     bool
	}{
		{
			name: "sucesso",
			body: map[string]string{
				"email":    "user@example.com",
				"name":     "Test User",
				"password": "secret123",
			},
			registerFn: func(_ context.Context, _ *auth.RegisterInput) (*auth.RegisterOutput, error) {
				return successOutput, nil
			},
			wantStatus: http.StatusCreated,
			wantToken:  true,
		},
		{
			name:          "JSON malformado",
			rawBody:       "{invalid",
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRequest.Error(),
		},
		{
			name:          "campo desconhecido no JSON",
			body:          map[string]string{"unknown": "field"},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRequest.Error(),
		},
		{
			name: "email já utilizado",
			body: map[string]string{
				"email":    "email@example.com",
				"name":     "Name",
				"password": "password123",
			},
			registerFn: func(_ context.Context, _ *auth.RegisterInput) (*auth.RegisterOutput, error) {
				return nil, coreerrors.ErrEmailAlreadyUsed
			},
			wantStatus:    http.StatusConflict,
			wantErrorCode: coreerrors.ErrEmailAlreadyUsed.Error(),
		},
		{
			name: "erro interno do servidor",
			body: map[string]string{
				"email":    "user@example.com",
				"name":     "Test User",
				"password": "secret123",
			},
			registerFn: func(_ context.Context, _ *auth.RegisterInput) (*auth.RegisterOutput, error) {
				return nil, errors.New("falha inesperada no banco de dados")
			},
			wantStatus:    http.StatusInternalServerError,
			wantErrorCode: coreerrors.ErrUnknown.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockAuthUsecase{registerFn: tc.registerFn}
			handler := auth.NewAuthHandler(mock)

			req := httptest.NewRequest(http.MethodPost, "/v1/auth/register", testutil.BuildBody(t, tc.body, tc.rawBody))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Register(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status HTTP: esperado %d, recebido %d", tc.wantStatus, rec.Code)
			}

			if tc.wantErrorCode != "" {
				resp := testutil.DecodeErrorResponse(t, rec.Body)
				if resp.ErrorCode != tc.wantErrorCode {
					t.Errorf("error_code: esperado %q, recebido %q", tc.wantErrorCode, resp.ErrorCode)
				}
			}

			if tc.wantToken {
				var resp auth.RegisterResponse
				if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
					t.Fatalf("erro ao decodificar resposta de sucesso: %v", err)
				}
				if resp.AccessToken == "" {
					t.Error("access_token não deveria estar vazio na resposta")
				}
				if resp.User.ID == uuid.Nil {
					t.Error("ID do usuário não deveria ser nulo na resposta")
				}
			}
		})
	}
}

func TestLoginHandler(t *testing.T) {
	successOutput := &auth.LoginOutput{
		User:        auth.UserOutput{ID: testutil.FixedUserID, Email: "user@example.com", Name: "Test User"},
		AccessToken: "token-login-ok",
	}

	tests := []struct {
		name          string
		body          any
		rawBody       string
		loginFn       func(context.Context, *auth.LoginInput) (*auth.LoginOutput, error)
		wantStatus    int
		wantErrorCode string
		wantToken     bool
	}{
		{
			name: "sucesso",
			body: map[string]string{
				"email":    "user@example.com",
				"password": "correct-password",
			},
			loginFn: func(_ context.Context, _ *auth.LoginInput) (*auth.LoginOutput, error) {
				return successOutput, nil
			},
			wantStatus: http.StatusOK,
			wantToken:  true,
		},
		{
			name:          "JSON malformado",
			rawBody:       "{invalid",
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRequest.Error(),
		},
		{
			name:          "campo desconhecido no JSON",
			body:          map[string]string{"unknown": "field"},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRequest.Error(),
		},
		{
			name: "usuário não encontrado",
			body: map[string]string{
				"email":    "ghost@example.com",
				"password": "whatever",
			},
			loginFn: func(_ context.Context, _ *auth.LoginInput) (*auth.LoginOutput, error) {
				return nil, coreerrors.ErrUserNotFound
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: coreerrors.ErrUserNotFound.Error(),
		},
		{
			name: "credenciais inválidas",
			body: map[string]string{
				"email":    "user@example.com",
				"password": "wrong-password",
			},
			loginFn: func(_ context.Context, _ *auth.LoginInput) (*auth.LoginOutput, error) {
				return nil, coreerrors.ErrInvalidCredentials
			},
			wantStatus:    http.StatusUnauthorized,
			wantErrorCode: coreerrors.ErrInvalidCredentials.Error(),
		},
		{
			name: "erro interno do servidor",
			body: map[string]string{
				"email":    "user@example.com",
				"password": "some-password",
			},
			loginFn: func(_ context.Context, _ *auth.LoginInput) (*auth.LoginOutput, error) {
				return nil, errors.New("falha inesperada")
			},
			wantStatus:    http.StatusInternalServerError,
			wantErrorCode: coreerrors.ErrUnknown.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockAuthUsecase{loginFn: tc.loginFn}
			handler := auth.NewAuthHandler(mock)

			req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", testutil.BuildBody(t, tc.body, tc.rawBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			if w.Code != tc.wantStatus {
				t.Errorf("status HTTP: esperado %d, recebido %d", tc.wantStatus, w.Code)
			}

			if tc.wantErrorCode != "" {
				resp := testutil.DecodeErrorResponse(t, w.Body)
				if resp.ErrorCode != tc.wantErrorCode {
					t.Errorf("error_code: esperado %q, recebido %q", tc.wantErrorCode, resp.ErrorCode)
				}
			}

			if tc.wantToken {
				var resp auth.LoginResponse
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("erro ao decodificar resposta de sucesso: %v", err)
				}
				if resp.AccessToken == "" {
					t.Error("access_token não deveria estar vazio na resposta")
				}
				if resp.User.ID == uuid.Nil {
					t.Error("ID do usuário não deveria ser nulo na resposta")
				}
			}
		})
	}
}
