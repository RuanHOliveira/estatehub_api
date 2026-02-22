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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/middlewares"
	"github.com/RuanHOliveira/estatehub_api/internal/core/testutil"
	"github.com/RuanHOliveira/estatehub_api/internal/domain/property_ads"
)

type mockPropertyAdUsecase struct {
	createFn func(ctx context.Context, input *property_ads.CreatePropertyAdInput) (*property_ads.CreatePropertyAdOutput, error)
	listFn   func(ctx context.Context) ([]property_ads.PropertyAdItem, error)
	deleteFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockPropertyAdUsecase) CreatePropertyAd(ctx context.Context, input *property_ads.CreatePropertyAdInput) (*property_ads.CreatePropertyAdOutput, error) {
	return m.createFn(ctx, input)
}

func (m *mockPropertyAdUsecase) ListPropertyAds(ctx context.Context) ([]property_ads.PropertyAdItem, error) {
	return m.listFn(ctx)
}

func (m *mockPropertyAdUsecase) DeletePropertyAd(ctx context.Context, id uuid.UUID) error {
	return m.deleteFn(ctx, id)
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

func buildDeleteRequest(t *testing.T, id string, withUserID bool) *http.Request {
	t.Helper()
	req := httptest.NewRequest(http.MethodDelete, "/v1/property-ads/"+id, nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)

	if withUserID {
		ctx = context.WithValue(ctx, middlewares.UserIDContextKey, testutil.FixedUserID)
	}

	return req.WithContext(ctx)
}

func fixedOutput(imagePath string) *property_ads.CreatePropertyAdOutput {
	var ip *string
	if imagePath != "" {
		ip = &imagePath
	}
	usd := 900000.0
	return &property_ads.CreatePropertyAdOutput{
		ID:           uuid.New(),
		UserID:       testutil.FixedUserID,
		Type:         "SALE",
		PriceBrl:     450000.00,
		PriceUsd:     &usd,
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
				if out.PriceUsd == nil {
					t.Error("price_usd não deveria ser nil quando usecase retorna cotação")
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

func TestPropertyAdHandler_ListPropertyAds(t *testing.T) {
	fixedID := uuid.New()
	fixedImagePath := "/uploads/property_ads/foto.jpg"
	usd := 900000.0

	tests := []struct {
		name          string
		listFn        func(ctx context.Context) ([]property_ads.PropertyAdItem, error)
		wantStatus    int
		wantErrorCode string
		checkResponse func(t *testing.T, body *bytes.Buffer)
	}{
		{
			name: "sucesso com resultados e price_usd",
			listFn: func(_ context.Context) ([]property_ads.PropertyAdItem, error) {
				return []property_ads.PropertyAdItem{
					{
						ID:        fixedID,
						UserID:    testutil.FixedUserID,
						Type:      "SALE",
						PriceBrl:  450000.00,
						PriceUsd:  &usd,
						ImagePath: &fixedImagePath,
						ZipCode:   "01310-100",
						Street:    "Av. Paulista",
						Number:    "1000",
						City:      "São Paulo",
						State:     "SP",
					},
				}, nil
			},
			wantStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body *bytes.Buffer) {
				var out []property_ads.PropertyAdItem
				if err := json.NewDecoder(body).Decode(&out); err != nil {
					t.Fatalf("erro ao decodificar resposta: %v", err)
				}
				if len(out) != 1 {
					t.Fatalf("esperava 1 item, recebeu %d", len(out))
				}
				if out[0].ID != fixedID {
					t.Errorf("id: esperado %v, recebido %v", fixedID, out[0].ID)
				}
				if out[0].Type != "SALE" {
					t.Errorf("type: esperado %q, recebido %q", "SALE", out[0].Type)
				}
				if out[0].PriceBrl != 450000.00 {
					t.Errorf("price_brl: esperado 450000.00, recebido %f", out[0].PriceBrl)
				}
				if out[0].PriceUsd == nil {
					t.Error("price_usd não deveria ser nil quando usecase retorna cotação")
				} else if *out[0].PriceUsd != usd {
					t.Errorf("price_usd: esperado %f, recebido %f", usd, *out[0].PriceUsd)
				}
				if out[0].ImagePath == nil || *out[0].ImagePath != fixedImagePath {
					t.Errorf("image_path: esperado %q", fixedImagePath)
				}
			},
		},
		{
			name: "sucesso lista vazia",
			listFn: func(_ context.Context) ([]property_ads.PropertyAdItem, error) {
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
			listFn: func(_ context.Context) ([]property_ads.PropertyAdItem, error) {
				return nil, errors.New("falha inesperada no banco")
			},
			wantStatus:    http.StatusInternalServerError,
			wantErrorCode: coreerrors.ErrUnknown.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockPropertyAdUsecase{listFn: tc.listFn}
			handler := property_ads.NewPropertyAdHandler(mock, t.TempDir())

			req := httptest.NewRequest(http.MethodGet, "/v1/property-ads", nil)
			rec := httptest.NewRecorder()
			handler.ListPropertyAds(rec, req)

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
				tc.checkResponse(t, bytes.NewBuffer(bodyBytes))
			}
		})
	}
}

func TestPropertyAdHandler_DeletePropertyAd(t *testing.T) {
	fixedID := uuid.New()

	tests := []struct {
		name          string
		req           *http.Request
		deleteFn      func(ctx context.Context, id uuid.UUID) error
		wantStatus    int
		wantErrorCode string
	}{
		{
			name:       "sucesso",
			req:        buildDeleteRequest(t, fixedID.String(), true),
			deleteFn:   func(_ context.Context, _ uuid.UUID) error { return nil },
			wantStatus: http.StatusOK,
		},
		{
			name:          "ID inválido",
			req:           buildDeleteRequest(t, "not-a-uuid", true),
			wantStatus:    http.StatusBadRequest,
			wantErrorCode: coreerrors.ErrInvalidRequest.Error(),
		},
		{
			name: "não encontrado",
			req:  buildDeleteRequest(t, fixedID.String(), true),
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return coreerrors.ErrPropertyAdNotFound
			},
			wantStatus:    http.StatusNotFound,
			wantErrorCode: coreerrors.ErrPropertyAdNotFound.Error(),
		},
		{
			name: "erro interno do usecase",
			req:  buildDeleteRequest(t, fixedID.String(), true),
			deleteFn: func(_ context.Context, _ uuid.UUID) error {
				return errors.New("falha inesperada no banco")
			},
			wantStatus:    http.StatusInternalServerError,
			wantErrorCode: coreerrors.ErrUnknown.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mock := &mockPropertyAdUsecase{deleteFn: tc.deleteFn}
			handler := property_ads.NewPropertyAdHandler(mock, t.TempDir())

			rec := httptest.NewRecorder()
			handler.DeletePropertyAd(rec, tc.req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: esperado %d, recebido %d", tc.wantStatus, rec.Code)
			}

			if tc.wantErrorCode != "" {
				errResp := testutil.DecodeErrorResponse(t, rec.Body)
				if errResp.ErrorCode != tc.wantErrorCode {
					t.Errorf("error_code: esperado %q, recebido %q", tc.wantErrorCode, errResp.ErrorCode)
				}
			}
		})
	}
}
