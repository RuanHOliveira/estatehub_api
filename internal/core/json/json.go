package json

import (
	"encoding/json"
	"net/http"

	errors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
)

// ErrorResponse é a resposta padrão para erros da API.
type ErrorResponse struct {
	ErrorCode string `json:"error_code" example:"ErrUnknown"`
}

func Read(r *http.Request, data any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}

func Write(w http.ResponseWriter, status int, data ...any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if len(data) == 0 {
		return
	}

	json.NewEncoder(w).Encode(data[0])
}

func WriteError(w http.ResponseWriter, status int, err error) {
	if err == nil {
		err = errors.ErrUnknown
	}

	Write(w, status, ErrorResponse{ErrorCode: err.Error()})
}
