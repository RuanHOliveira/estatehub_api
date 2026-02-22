package exchange_rates_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/middlewares"
	"github.com/RuanHOliveira/estatehub_api/internal/core/testutil"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/exchange_rates"
	"github.com/google/uuid"
)

type mockExchangeRateUsecase struct {
	createFn func(ctx context.Context, input *exchange_rates.CreateExchangeRateInput) (*exchange_rates.CreateExchangeRateOutput, error)
	listFn   func(ctx context.Context) ([]exchange_rates.ExchangeRateItem, error)
}

func (m *mockExchangeRateUsecase) CreateExchangeRate(ctx context.Context, input *exchange_rates.CreateExchangeRateInput) (*exchange_rates.CreateExchangeRateOutput, error) {
	return m.createFn(ctx, input)
}

func (m *mockExchangeRateUsecase) ListAllExchangeRates(ctx context.Context) ([]exchange_rates.ExchangeRateItem, error) {
	return m.listFn(ctx)
}

func buildAuthRequest(t *testing.T, body any, rawBody string, withAuth bool) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/v1/exchange-rates", testutil.BuildBody(t, body, rawBody))
	req.Header.Set("Content-Type", "application/json")
	if withAuth {
		ctx := context.WithValue(req.Context(), middlewares.UserIDContextKey, testutil.FixedUserID)
		req = req.WithContext(ctx)
	}
	return req
}

func fixedOutput() *exchange_rates.CreateExchangeRateOutput {
	return &exchange_rates.CreateExchangeRateOutput{
		ID:             uuid.New(),
		UserID:         testutil.FixedUserID,
		TargetCurrency: "USD",
		Rate:           2.50,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func TestExchangeRateHandler_CreateExchangeRate(t *testing.T) {
	validBody := map[string]string{
		"target_currency": "USD",
		"rate":            "2.50",
	}

	tests := []struct {
		name          string
		req           *http.Request
		createFn      func(ctx context.Context, input *exchange_rates.CreateExchangeRateInput) (*exchange_rates.CreateExchangeRateOutput, error)
		wantStatus    int
		wantErrorCode string
		checkResponse func(t *testing.T, body *bytes.Buffer)
	}{
		{
			name: "sucesso",
			req:  buildAuthRequest(t, validBody, "", true),
			createFn: func(_ context.Context, _ *exchange_rates.CreateExchangeRateInput) (*exchange_rates.CreateExchangeRateOutput, error) {
				return fixedOutput(), nil
			},
			wantStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body *bytes.Buffer) {
				var out exchange_rates.CreateExchangeRateOutput
				if err := json.NewDecoder(body).Decode(&out); err != nil {
					t.Fatalf("erro ao decodificar resposta: %v", err)
				}
				if out.ID == uuid.Nil {
					t.Error("ID da cotação não deveria ser nulo")
				}
				if out.TargetCurrency != "USD" {
					t.Errorf("target_currency: esperado %q, recebido %q", "USD", out.TargetCurrency)
				}
				if out.Rate != 2.50 {
					t.Errorf("rate: esperado 2.50, recebido %f", out.Rate)
				}
			},
		},
		{
			name:          "JSON malformado",
			req:           buildAuthRequest(t, nil, "{invalid", true),
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRequest.Error(),
		},
		{
			name:          "campo desconhecido no JSON",
			req:           buildAuthRequest(t, map[string]string{"unknown": "x"}, "", true),
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRequest.Error(),
		},
		{
			name:          "não autenticado",
			req:           buildAuthRequest(t, validBody, "", false),
			wantStatus:    http.StatusUnauthorized,
			wantErrorCode: coreerrors.ErrMissingToken.Error(),
		},
		{
			name: "rate não numérica",
			req: buildAuthRequest(t, map[string]string{
				"target_currency": "USD",
				"rate":            "abc",
			}, "", true),
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRequest.Error(),
		},
		{
			name: "usecase retorna ErrInvalidRate",
			req:  buildAuthRequest(t, validBody, "", true),
			createFn: func(_ context.Context, _ *exchange_rates.CreateExchangeRateInput) (*exchange_rates.CreateExchangeRateOutput, error) {
				return nil, coreerrors.ErrInvalidRate
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRate.Error(),
		},
		{
			name: "erro interno do usecase",
			req:  buildAuthRequest(t, validBody, "", true),
			createFn: func(_ context.Context, _ *exchange_rates.CreateExchangeRateInput) (*exchange_rates.CreateExchangeRateOutput, error) {
				return nil, errors.New("falha inesperada no banco")
			},
			wantStatus:    http.StatusInternalServerError,
			wantErrorCode: coreerrors.ErrUnknown.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockExchangeRateUsecase{createFn: tc.createFn}
			handler := exchange_rates.NewExchangeRateHandler(mock)

			rec := httptest.NewRecorder()
			handler.CreateExchangeRate(rec, tc.req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status HTTP: esperado %d, recebido %d", tc.wantStatus, rec.Code)
			}

			if tc.wantErrorCode != "" {
				resp := testutil.DecodeErrorResponse(t, rec.Body)
				if resp.ErrorCode != tc.wantErrorCode {
					t.Errorf("error_code: esperado %q, recebido %q", tc.wantErrorCode, resp.ErrorCode)
				}
				return
			}

			if tc.checkResponse != nil {
				bodyBytes := rec.Body.Bytes()
				tc.checkResponse(t, bytes.NewBuffer(bodyBytes))
			}
		})
	}
}

func TestExchangeRateHandler_ListAllExchangeRates(t *testing.T) {
	fixedID := uuid.New()
	deletedID := uuid.New()
	deletedAt := time.Now()

	tests := []struct {
		name          string
		listFn        func(ctx context.Context) ([]exchange_rates.ExchangeRateItem, error)
		wantStatus    int
		wantErrorCode string
		checkResponse func(t *testing.T, body *bytes.Buffer)
	}{
		{
			name: "sucesso incluindo registros deletados",
			listFn: func(_ context.Context) ([]exchange_rates.ExchangeRateItem, error) {
				return []exchange_rates.ExchangeRateItem{
					{
						ID:             fixedID,
						UserID:         testutil.FixedUserID,
						TargetCurrency: "USD",
						Rate:           2.50,
						CreatedAt:      time.Now(),
						UpdatedAt:      time.Now(),
						DeletedAt:      nil,
					},
					{
						ID:             deletedID,
						UserID:         testutil.FixedUserID,
						TargetCurrency: "USD",
						Rate:           5.00,
						CreatedAt:      time.Now(),
						UpdatedAt:      time.Now(),
						DeletedAt:      &deletedAt,
					},
				}, nil
			},
			wantStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body *bytes.Buffer) {
				var out []map[string]any
				if err := json.NewDecoder(body).Decode(&out); err != nil {
					t.Fatalf("erro ao decodificar resposta: %v", err)
				}
				if len(out) != 2 {
					t.Fatalf("esperava 2 itens (incluindo deletado), recebeu %d", len(out))
				}
				if out[0]["deleted_at"] != nil {
					t.Errorf("primeiro item deveria ter deleted_at null, recebeu %v", out[0]["deleted_at"])
				}
				if out[1]["deleted_at"] == nil {
					t.Error("segundo item deveria ter deleted_at preenchido")
				}
			},
		},
		{
			name: "sucesso lista vazia",
			listFn: func(_ context.Context) ([]exchange_rates.ExchangeRateItem, error) {
				return nil, nil
			},
			wantStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body *bytes.Buffer) {
				var raw json.RawMessage
				if err := json.NewDecoder(body).Decode(&raw); err != nil {
					t.Fatalf("erro ao decodificar resposta: %v", err)
				}
				if string(raw) != "[]" {
					t.Errorf("esperava [], recebeu %s", string(raw))
				}
			},
		},
		{
			name: "erro do usecase",
			listFn: func(_ context.Context) ([]exchange_rates.ExchangeRateItem, error) {
				return nil, errors.New("falha inesperada no banco")
			},
			wantStatus:    http.StatusInternalServerError,
			wantErrorCode: coreerrors.ErrUnknown.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockExchangeRateUsecase{listFn: tc.listFn}
			handler := exchange_rates.NewExchangeRateHandler(mock)

			req := httptest.NewRequest(http.MethodGet, "/v1/exchange-rates", nil)
			rec := httptest.NewRecorder()
			handler.ListAllExchangeRates(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status HTTP: esperado %d, recebido %d", tc.wantStatus, rec.Code)
			}

			if tc.wantErrorCode != "" {
				resp := testutil.DecodeErrorResponse(t, rec.Body)
				if resp.ErrorCode != tc.wantErrorCode {
					t.Errorf("error_code: esperado %q, recebido %q", tc.wantErrorCode, resp.ErrorCode)
				}
				return
			}

			if tc.checkResponse != nil {
				bodyBytes := rec.Body.Bytes()
				tc.checkResponse(t, bytes.NewBuffer(bodyBytes))
			}
		})
	}
}
