package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	corejson "github.com/RuanHOliveira/estatehub_api/internal/core/json"
	"github.com/google/uuid"
)

var FixedUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

func BuildBody(t *testing.T, body any, rawBody string) io.Reader {
	t.Helper()
	if rawBody != "" {
		return strings.NewReader(rawBody)
	}
	if body == nil {
		return http.NoBody
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("erro ao serializar corpo da requisição: %v", err)
	}
	return bytes.NewReader(b)
}

func DecodeErrorResponse(t *testing.T, body *bytes.Buffer) corejson.ErrorResponse {
	t.Helper()
	var resp corejson.ErrorResponse
	if err := json.NewDecoder(body).Decode(&resp); err != nil {
		t.Fatalf("erro ao decodificar resposta de erro: %v", err)
	}
	return resp
}
