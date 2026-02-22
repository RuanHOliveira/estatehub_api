package exchange_rates

import (
	"log"
	"net/http"
	"strconv"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/json"
	"github.com/RuanHOliveira/estatehub_api/internal/core/middlewares"
	"github.com/google/uuid"
)

type ExchangeRateHandler struct {
	u ExchangeRateUsecase
}

func NewExchangeRateHandler(u ExchangeRateUsecase) *ExchangeRateHandler {
	return &ExchangeRateHandler{u: u}
}

// CreateExchangeRate godoc
// @Summary      Criar cotação de câmbio
// @Description  Registra nova cotação BRL→USD. Inativa automaticamente todas as cotações anteriores.
// @Tags         exchange-rates
// @Accept       json
// @Produce      json
// @Param        request body CreateExchangeRateRequest true "Dados da cotação"
// @Success      201 {object} CreateExchangeRateOutput
// @Failure      400 {object} json.ErrorResponse
// @Failure      401 {object} json.ErrorResponse
// @Failure      500 {object} json.ErrorResponse
// @Security     BearerAuth
// @Router       /exchange-rates [post]
func (h *ExchangeRateHandler) CreateExchangeRate(w http.ResponseWriter, r *http.Request) {
	var req CreateExchangeRateRequest
	if err := json.Read(r, &req); err != nil {
		log.Println(err)
		json.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidRequest)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDContextKey).(uuid.UUID)
	if !ok {
		json.WriteError(w, http.StatusUnauthorized, coreerrors.ErrMissingToken)
		return
	}

	rate, err := strconv.ParseFloat(req.RateStr, 64)
	if err != nil {
		json.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidRequest)
		return
	}

	output, err := h.u.CreateExchangeRate(r.Context(), &CreateExchangeRateInput{UserID: userID, TargetCurrency: req.TargetCurrency, Rate: rate})
	if err != nil {
		log.Println(err)
		switch err {
		case coreerrors.ErrInvalidRate:
			json.WriteError(w, http.StatusBadRequest, err)
		default:
			json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
		}
		return
	}

	json.Write(w, http.StatusCreated, output)
}

// ListAllExchangeRates godoc
// @Summary      Listar cotações de câmbio
// @Description  Retorna o histórico completo de cotações, incluindo inativas (deleted_at != null). A cotação ativa possui deleted_at nulo.
// @Tags         exchange-rates
// @Produce      json
// @Success      200 {array}  ExchangeRateItem
// @Failure      401 {object} json.ErrorResponse
// @Failure      500 {object} json.ErrorResponse
// @Security     BearerAuth
// @Router       /exchange-rates [get]
func (h *ExchangeRateHandler) ListAllExchangeRates(w http.ResponseWriter, r *http.Request) {
	ers, err := h.u.ListAllExchangeRates(r.Context())
	if err != nil {
		log.Println(err)
		json.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
		return
	}

	if ers == nil {
		ers = []ExchangeRateItem{}
	}

	json.Write(w, http.StatusOK, ers)
}
