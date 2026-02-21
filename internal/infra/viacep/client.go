package viacep

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	errors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
)

type ViaCEPClient interface {
	FindAddressByCEP(ctx context.Context, cep string) (ViaCEPAddress, error)
}

// Valida que o CEP contém exatamente 8 dígitos numéricos.
var cepRegex = regexp.MustCompile(`^\d{8}$`)

const (
	viaCEPBaseURL  = "https://viacep.com.br/ws"
	defaultTimeout = 5 * time.Second
)

type httpViaCEPClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewViaCEPClient() ViaCEPClient {
	return &httpViaCEPClient{
		httpClient: &http.Client{Timeout: defaultTimeout},
		baseURL:    viaCEPBaseURL,
	}
}

func NewViaCEPClientWithConfig(baseURL string, timeout time.Duration) ViaCEPClient {
	return &httpViaCEPClient{
		httpClient: &http.Client{Timeout: timeout},
		baseURL:    baseURL,
	}
}

func (c *httpViaCEPClient) FindAddressByCEP(ctx context.Context, cep string) (ViaCEPAddress, error) {
	if !cepRegex.MatchString(cep) {
		return ViaCEPAddress{}, errors.ErrInvalidCEP
	}

	url := fmt.Sprintf("%s/%s/json/", c.baseURL, cep)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ViaCEPAddress{}, errors.ErrExternalServiceFailure
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ViaCEPAddress{}, errors.ErrExternalServiceFailure
	}
	defer resp.Body.Close()

	// ViaCEP retorna 400 para CEPs com formato inválido na URL
	if resp.StatusCode == http.StatusBadRequest {
		return ViaCEPAddress{}, errors.ErrExternalBadRequest
	}

	if resp.StatusCode != http.StatusOK {
		return ViaCEPAddress{}, errors.ErrExternalServiceFailure
	}

	// ViaCEP retorna 200 com {"erro": true} quando o CEP não existe
	var viaCEP viaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&viaCEP); err != nil {
		return ViaCEPAddress{}, errors.ErrCEPNotFound
	}

	return ViaCEPAddress{
		ZipCode:      viaCEP.CEP,
		Street:       viaCEP.Logradouro,
		Neighborhood: viaCEP.Bairro,
		City:         viaCEP.Localidade,
		State:        viaCEP.UF,
		Complement:   viaCEP.Complemento,
	}, nil
}
