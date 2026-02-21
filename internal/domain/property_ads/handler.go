package property_ads

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	corejson "github.com/RuanHOliveira/estatehub_api/internal/core/json"
	"github.com/RuanHOliveira/estatehub_api/internal/core/middlewares"
	"github.com/google/uuid"
)

const maxImageSize = 5 << 20 // 5MB

type PropertyAdHandler struct {
	u         PropertyAdUsecase
	uploadDir string
}

func NewPropertyAdHandler(u PropertyAdUsecase, uploadDir string) *PropertyAdHandler {
	return &PropertyAdHandler{u: u, uploadDir: uploadDir}
}

func (h *PropertyAdHandler) CreatePropertyAd(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middlewares.UserIDContextKey).(uuid.UUID)
	if !ok {
		corejson.WriteError(w, http.StatusUnauthorized, coreerrors.ErrMissingToken)
		return
	}

	// Limita tamanho do body e parsear multipart form
	r.Body = http.MaxBytesReader(w, r.Body, maxImageSize+4096)
	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		log.Println(err)
		corejson.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidRequest)
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
		corejson.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidPrice)
		return
	}

	// Processa imagem
	imagePath := ""
	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()

		// Verifica tamanho
		if header.Size > maxImageSize {
			corejson.WriteError(w, http.StatusBadRequest, coreerrors.ErrImageTooLarge)
			return
		}

		// Pega Content-Type
		buf := make([]byte, 512)
		n, err := file.Read(buf)
		if err != nil && err != io.EOF {
			log.Println(err)
			corejson.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}
		contentType := http.DetectContentType(buf[:n])
		if contentType != "image/jpeg" && contentType != "image/png" {
			corejson.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidImageType)
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
			corejson.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}

		// Cria arquivo de destino
		fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
		destPath := filepath.Join(h.uploadDir, fileName)
		dst, err := os.Create(destPath)
		if err != nil {
			log.Println(err)
			corejson.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}
		defer dst.Close()

		if _, err := dst.Write(buf[:n]); err != nil {
			log.Println(err)
			corejson.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}
		if _, err := io.Copy(dst, file); err != nil {
			log.Println(err)
			corejson.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
			return
		}

		imagePath = fmt.Sprintf("/uploads/property_ads/%s", fileName)
	} else if err != http.ErrMissingFile {
		log.Println(err)
		corejson.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidRequest)
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
			corejson.WriteError(w, http.StatusBadRequest, err)
		default:
			corejson.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
		}
		return
	}

	corejson.Write(w, http.StatusCreated, output)
}
