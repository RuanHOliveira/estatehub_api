package viacep_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/infra/viacep"
)

const respostaSucesso = `{
	"cep": "01001-000",
	"logradouro": "Praça da Sé",
	"complemento": "lado ímpar",
	"bairro": "Sé",
	"localidade": "São Paulo",
	"uf": "SP",
	"estado": "São Paulo",
	"regiao": "Sudeste",
	"ibge": "3550308",
	"gia": "1004",
	"ddd": "11",
	"siafi": "7107"
}`

func TestViaCEPClient_FindAddressByCEP(t *testing.T) {
	tests := []struct {
		name        string
		cep         string
		handler     http.HandlerFunc
		timeout     time.Duration
		wantErr     error
		wantZipCode string
		wantCity    string
		wantState   string
	}{
		{
			name: "sucesso",
			cep:  "01001000",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(respostaSucesso))
			},
			wantZipCode: "01001-000",
			wantCity:    "São Paulo",
			wantState:   "SP",
		},
		{
			// CEP com formato inválido não chega a fazer chamada HTTP
			name:    "CEP inválido",
			cep:     "123",
			handler: nil,
			wantErr: coreerrors.ErrInvalidCEP,
		},
		{
			// ViaCEP retorna HTTP 400 para CEPs malformados na URL
			name: "HTTP 400 do ViaCEP",
			cep:  "00000000",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			wantErr: coreerrors.ErrExternalBadRequest,
		},
		{
			// ViaCEP retorna 200 com {"erro": true} quando o CEP não existe
			name: "CEP não encontrado",
			cep:  "00000000",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"erro": true}`))
			},
			wantErr: coreerrors.ErrCEPNotFound,
		},
		{
			// Qualquer status HTTP inesperado resulta em falha externa
			name: "HTTP inesperado do servidor",
			cep:  "01001000",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: coreerrors.ErrExternalServiceFailure,
		},
		{
			// O server demora mais do que o timeout configurado no client
			name:    "timeout de rede",
			cep:     "01001000",
			timeout: 1 * time.Millisecond,
			handler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond)
				w.WriteHeader(http.StatusOK)
			},
			wantErr: coreerrors.ErrExternalServiceFailure,
		},
		{
			// Resposta com corpo JSON inválido
			name: "JSON inválido na resposta",
			cep:  "01001000",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{invalid`))
			},
			wantErr: coreerrors.ErrInvalidExternalResponse,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.handler == nil {
				client := viacep.NewViaCEPClientWithConfig("http://localhost", 5*time.Second)
				_, err := client.FindAddressByCEP(context.Background(), tc.cep)
				if !errors.Is(err, tc.wantErr) {
					t.Errorf("erro esperado %v, recebido %v", tc.wantErr, err)
				}
				return
			}

			ts := httptest.NewServer(tc.handler)
			defer ts.Close()

			timeout := tc.timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			client := viacep.NewViaCEPClientWithConfig(ts.URL, timeout)

			address, err := client.FindAddressByCEP(context.Background(), tc.cep)

			if tc.wantErr == nil {
				if err != nil {
					t.Fatalf("não esperava erro, recebido: %v", err)
				}
				if address.ZipCode != tc.wantZipCode {
					t.Errorf("ZipCode: esperado %q, recebido %q", tc.wantZipCode, address.ZipCode)
				}
				if address.City != tc.wantCity {
					t.Errorf("City: esperado %q, recebido %q", tc.wantCity, address.City)
				}
				if address.State != tc.wantState {
					t.Errorf("State: esperado %q, recebido %q", tc.wantState, address.State)
				}
				return
			}

			if !errors.Is(err, tc.wantErr) {
				t.Errorf("erro esperado %v, recebido %v", tc.wantErr, err)
			}
		})
	}
}
