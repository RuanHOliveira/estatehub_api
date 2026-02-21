package property_ads_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"path/filepath"
	"testing"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/middlewares"
	"github.com/RuanHOliveira/estatehub_api/internal/core/testutil"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/property_ads"
	"github.com/google/uuid"
)

type mockPropertyAdUsecase struct {
	createFn func(ctx context.Context, input *property_ads.CreatePropertyAdInput) (*property_ads.CreatePropertyAdOutput, error)
}

func (m *mockPropertyAdUsecase) CreatePropertyAd(ctx context.Context, input *property_ads.CreatePropertyAdInput) (*property_ads.CreatePropertyAdOutput, error) {
	return m.createFn(ctx, input)
}

func validFields() map[string]string {
	return map[string]string{
		"type":         "SALE",
		"price_brl":    "450000.00",
		"zip_code":     "01310-100",
		"street":       "Av. Paulista",
		"number":       "1000",
		"neighborhood": "Bela Vista",
		"city":         "São Paulo",
		"state":        "SP",
	}
}

func buildMultipartRequest(
	t *testing.T,
	fields map[string]string,
	withUserID bool,
	imageData []byte,
	imageContentType string,
) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)

	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatalf("erro ao escrever campo %q: %v", k, err)
		}
	}

	if imageData != nil {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="image"; filename="test.img"`)
		h.Set("Content-Type", imageContentType)
		part, err := mw.CreatePart(h)
		if err != nil {
			t.Fatalf("erro ao criar parte da imagem: %v", err)
		}
		if _, err := part.Write(imageData); err != nil {
			t.Fatalf("erro ao escrever dados da imagem: %v", err)
		}
	}

	mw.Close()

	req := httptest.NewRequest(http.MethodPost, "/v1/property-ads", &buf)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	if withUserID {
		ctx := context.WithValue(req.Context(), middlewares.UserIDContextKey, testutil.FixedUserID)
		req = req.WithContext(ctx)
	}

	return req
}

func fixedOutput(imagePath string) *property_ads.CreatePropertyAdOutput {
	var ip *string
	if imagePath != "" {
		ip = &imagePath
	}
	return &property_ads.CreatePropertyAdOutput{
		ID:           uuid.New(),
		UserID:       testutil.FixedUserID,
		Type:         "SALE",
		PriceBrl:     450000.00,
		ImagePath:    ip,
		ZipCode:      "01310-100",
		Street:       "Av. Paulista",
		Number:       "1000",
		Neighborhood: "Bela Vista",
		City:         "São Paulo",
		State:        "SP",
	}
}

func TestPropertyAdHandler_CreatePropertyAd(t *testing.T) {
	tests := []struct {
		name          string
		req           func(uploadDir string) *http.Request
		createFn      func(ctx context.Context, input *property_ads.CreatePropertyAdInput) (*property_ads.CreatePropertyAdOutput, error)
		wantStatus    int
		wantErrorCode string
		checkResponse func(t *testing.T, body *bytes.Buffer, uploadDir string)
	}{
		{
			name: "sucesso sem imagem",
			req: func(_ string) *http.Request {
				return buildMultipartRequest(t, validFields(), true, nil, "")
			},
			createFn: func(_ context.Context, _ *property_ads.CreatePropertyAdInput) (*property_ads.CreatePropertyAdOutput, error) {
				return fixedOutput(""), nil
			},
			wantStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body *bytes.Buffer, _ string) {
				var out property_ads.CreatePropertyAdOutput
				if err := json.NewDecoder(body).Decode(&out); err != nil {
					t.Fatalf("erro ao decodificar resposta: %v", err)
				}
				if out.ID == uuid.Nil {
					t.Error("ID do anúncio não deveria ser nulo")
				}
				if out.ImagePath != nil {
					t.Errorf("image_path deveria ser nil quando sem imagem, recebeu %q", *out.ImagePath)
				}
			},
		},
		{
			name: "sucesso com imagem JPEG",
			req: func(_ string) *http.Request {
				return buildMultipartRequest(t, validFields(), true, testutil.JpegMagicBytes(), "image/jpeg")
			},
			createFn: func(_ context.Context, input *property_ads.CreatePropertyAdInput) (*property_ads.CreatePropertyAdOutput, error) {
				return fixedOutput(input.ImagePath), nil
			},
			wantStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, _ *bytes.Buffer, uploadDir string) {
				// Verifica se ao menos um arquivo .jpg foi criado no uploadDir
				matches, _ := filepath.Glob(filepath.Join(uploadDir, "*.jpg"))
				if len(matches) == 0 {
					t.Error("esperava arquivo .jpg salvo no uploadDir")
				}
			},
		},
		{
			name: "não autenticado",
			req: func(_ string) *http.Request {
				return buildMultipartRequest(t, validFields(), false, nil, "")
			},
			createFn:      nil,
			wantStatus:    http.StatusUnauthorized,
			wantErrorCode: coreerrors.ErrMissingToken.Error(),
		},
		{
			name: "price_brl inválido",
			req: func(_ string) *http.Request {
				fields := validFields()
				fields["price_brl"] = "abc"
				return buildMultipartRequest(t, fields, true, nil, "")
			},
			createFn:      nil,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidPrice.Error(),
		},
		{
			name: "body muito grande",
			req: func(_ string) *http.Request {
				// Criar dados maiores que 5MB + 4KB
				largeData := make([]byte, (5<<20)+4096+1)
				copy(largeData, testutil.JpegMagicBytes())
				return buildMultipartRequest(t, validFields(), true, largeData, "image/jpeg")
			},
			createFn:      nil,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRequest.Error(),
		},
		{
			name: "tipo de imagem inválido",
			req: func(_ string) *http.Request {
				return buildMultipartRequest(t, validFields(), true, testutil.GifBytes(), "image/gif")
			},
			createFn:      nil,
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidImageType.Error(),
		},
		{
			name: "usecase retorna ErrInvalidAdType",
			req: func(_ string) *http.Request {
				return buildMultipartRequest(t, validFields(), true, nil, "")
			},
			createFn: func(_ context.Context, _ *property_ads.CreatePropertyAdInput) (*property_ads.CreatePropertyAdOutput, error) {
				return nil, coreerrors.ErrInvalidAdType
			},
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidAdType.Error(),
		},
		{
			name: "usecase retorna erro interno",
			req: func(_ string) *http.Request {
				return buildMultipartRequest(t, validFields(), true, nil, "")
			},
			createFn: func(_ context.Context, _ *property_ads.CreatePropertyAdInput) (*property_ads.CreatePropertyAdOutput, error) {
				return nil, errors.New("falha inesperada no banco")
			},
			wantStatus:    http.StatusInternalServerError,
			wantErrorCode: coreerrors.ErrUnknown.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			uploadDir := t.TempDir()
			mock := &mockPropertyAdUsecase{createFn: tc.createFn}
			handler := property_ads.NewPropertyAdHandler(mock, uploadDir)

			req := tc.req(uploadDir)
			rec := httptest.NewRecorder()
			handler.CreatePropertyAd(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: esperado %d, recebido %d", tc.wantStatus, rec.Code)
			}

			if tc.wantErrorCode != "" {
				errResp := testutil.DecodeErrorResponse(t, rec.Body)
				if errResp.ErrorCode != tc.wantErrorCode {
					t.Errorf("error_code: esperado %q, recebido %q", tc.wantErrorCode, errResp.ErrorCode)
				}
				return
			}

			if tc.checkResponse != nil {
				bodyBytes := rec.Body.Bytes()
				tc.checkResponse(t, bytes.NewBuffer(bodyBytes), uploadDir)
			}

		})
	}
}
