package property_ads

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/json"
	"github.com/RuanHOliveira/estatehub_api/internal/core/middlewares"
)

const maxImageSize = 5 << 20 // 5MB

type PropertyAdHandler struct {
	u         PropertyAdUsecase
	uploadDir string
}

func NewPropertyAdHandler(u PropertyAdUsecase, uploadDir string) *PropertyAdHandler {
	return &PropertyAdHandler{u: u, uploadDir: uploadDir}
}

// CreatePropertyAd godoc
// @Summary      Criar anúncio imobiliário
// @Description  Cria um novo anúncio com upload opcional de imagem (JPEG/PNG, máx 5MB)
// @Tags         property-ads
// @Accept       multipart/form-data
// @Produce      json
// @Param        type         formData string  true  "Tipo do anúncio: SALE ou RENT"
// @Param        price_brl    formData number  true  "Preço em BRL (deve ser > 0)"
// @Param        zip_code     formData string  true  "CEP (8 dígitos)"
// @Param        street       formData string  true  "Logradouro"
// @Param        number       formData string  true  "Número"
// @Param        neighborhood formData string  true  "Bairro"
// @Param        city         formData string  true  "Cidade"
// @Param        state        formData string  true  "Estado (UF, 2 letras)"
// @Param        complement   formData string  false "Complemento"
// @Param        image        formData file    false "Imagem do imóvel (JPEG ou PNG, máx 5MB)"
// @Success      201 {object} CreatePropertyAdOutput
// @Failure      400 {object} json.ErrorResponse
// @Failure      401 {object} json.ErrorResponse
// @Failure      500 {object} json.ErrorResponse
// @Security     BearerAuth
// @Router       /property-ads [post]
func (h *PropertyAdHandler) CreatePropertyAd(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDContextKey).(uuid.UUID)
	if !ok {
		json.WriteError(w, http.StatusUnauthorized, coreerrors.ErrMissingToken)
		return
	}

	// Limita tamanho do body e parsear multipart form
	r.Body = http.MaxBytesReader(w, r.Body, maxImageSize+4096)
	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		log.Println(err)
		json.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidRequest)
		return
	}

	// Extrai campos
	req := CreatePropertyAdRequest{
		Type:         r.FormValue("type"),
		PriceBrlStr:  r.FormValue("price_brl"),
		ZipCode:      r.FormValue("zip_code"),
		Street:       r.FormValue("street"),
		Number:       r.FormValue("number"),
		Neighborhood: r.FormValue("neighborhood"),
		City:         r.FormValue("city"),
		State:        r.FormValue("state"),
		Complement:   r.FormValue("complement"),
	}

	// Converte price_brl para float64
	priceBrl, err := strconv.ParseFloat(req.PriceBrlStr, 64)
	if err != nil {
		json.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidPrice)
		return
	}

	// Processa imagem
	imagePath := ""
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Verifica tamanho
		if header.Size > maxImageSize {
			json.WriteError(w, http.StatusBadRequest, coreerrors.ErrImageTooLarge)
			return
		}

		// Pega Content-Type
		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			log.Println(err)
			json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}
		contentType := http.DetectContentType(buf[:n])
		if contentType != "image/jpeg" && contentType != "image/png" {
			json.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidImageType)
			return
		}

		// Determinar extensão pelo content-type detectado
		ext := ".jpg"
		if contentType == "image/png" {
			ext = ".png"
		}

		// Garante que o diretório de upload existe
		if err := os.MkdirAll(h.uploadDir, 0755); err != nil {
			log.Println(err)
			json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}

		// Cria arquivo de destino
		fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		destPath := filepath.Join(h.uploadDir, fileName)
		dst, err := os.Create(destPath)
		if err != nil {
			log.Println(err)
			json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}
		defer dst.Close()

		if _, err := dst.Write(buf[:n]); err != nil {
			log.Println(err)
			json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}
		if _, err := io.Copy(dst, file); err != nil {
			log.Println(err)
			json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}

		imagePath = fmt.Sprintf("/uploads/property_ads/%s", fileName)
	} else if err != http.ErrMissingFile {
		log.Println(err)
		json.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidRequest)
		return
	}

	output, err := h.u.CreatePropertyAd(r.Context(), &CreatePropertyAdInput{
		UserID:       userID,
		Type:         req.Type,
		PriceBrl:     priceBrl,
		ImagePath:    imagePath,
		ZipCode:      req.ZipCode,
		Street:       req.Street,
		Number:       req.Number,
		Neighborhood: req.Neighborhood,
		City:         req.City,
		State:        req.State,
		Complement:   req.Complement,
	})
	if err != nil {
		log.Println(err)
		switch err {
		case coreerrors.ErrInvalidAdType, coreerrors.ErrInvalidPrice, coreerrors.ErrMissingAddressField:
			json.WriteError(w, http.StatusBadRequest, err)
		default:
			json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
		}
		return
	}

	if output.ImagePath != nil {
		filePath := filepath.Join(h.uploadDir, filepath.Base(*output.ImagePath))
		if data, err := os.ReadFile(filePath); err == nil {
			ct := http.DetectContentType(data)
			encoded := "data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(data)
			output.ImageData = &encoded
		}
	}

	json.Write(w, http.StatusCreated, output)
}

// ListPropertyAds godoc
// @Summary      Listar anúncios imobiliários
// @Description  Retorna todos os anúncios ativos (soft delete não incluído). price_usd é nulo se não houver cotação ativa.
// @Tags         property-ads
// @Produce      json
// @Success      200 {array}  PropertyAdItem
// @Failure      401 {object} json.ErrorResponse
// @Failure      500 {object} json.ErrorResponse
// @Security     BearerAuth
// @Router       /property-ads [get]
func (h *PropertyAdHandler) ListPropertyAds(w http.ResponseWriter, r *http.Request) {
	ads, err := h.u.ListPropertyAds(r.Context())
	if err != nil {
		log.Println(err)
		json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
		return
	}

	if ads == nil {
		ads = []PropertyAdItem{}
	}

	for i, ad := range ads {
		if ad.ImagePath == nil {
			continue
		}

		filePath := filepath.Join(h.uploadDir, filepath.Base(*ad.ImagePath))
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Println(err)
			continue
		}
		contentType := http.DetectContentType(data)
		encoded := "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data)
		ads[i].ImageData = &encoded
	}

	json.Write(w, http.StatusOK, ads)
}

// DeletePropertyAd godoc
// @Summary      Deletar anúncio imobiliário
// @Description  Realiza soft delete do anúncio. Não remove o registro fisicamente do banco.
// @Tags         property-ads
// @Produce      json
// @Param        id   path string true "UUID do anúncio" Format(uuid)
// @Success      200
// @Failure      400 {object} json.ErrorResponse
// @Failure      401 {object} json.ErrorResponse
// @Failure      404 {object} json.ErrorResponse
// @Failure      500 {object} json.ErrorResponse
// @Security     BearerAuth
// @Router       /property-ads/{id} [delete]
func (h *PropertyAdHandler) DeletePropertyAd(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		json.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidRequest)
		return
	}

	if err := h.u.DeletePropertyAd(r.Context(), id); err != nil {
		log.Println(err)
		switch err {
		case coreerrors.ErrPropertyAdNotFound:
			json.WriteError(w, http.StatusNotFound, err)
		default:
			json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}
