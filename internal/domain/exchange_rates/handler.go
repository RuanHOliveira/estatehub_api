package exchange_rates

import (
	"log"
	"net/http"
	"strconv"

	coreerrors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	corejson "github.com/RuanHOliveira/estatehub_api/internal/core/json"
	"github.com/RuanHOliveira/estatehub_api/internal/core/middlewares"
	"github.com/google/uuid"
)

type ExchangeRateHandler struct {
	u ExchangeRateUsecase
}

func NewExchangeRateHandler(u ExchangeRateUsecase) *ExchangeRateHandler {
	return &ExchangeRateHandler{u: u}
}

func (h *ExchangeRateHandler) CreateExchangeRate(w http.ResponseWriter, r *http.Request) {
	var req CreateExchangeRateRequest
	if err := corejson.Read(r, &req); err != nil {
		log.Println(err)
		corejson.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidRequest)
		return
	}

	userID, ok := r.Context().Value(middlewares.UserIDContextKey).(uuid.UUID)
	if !ok {
		corejson.WriteError(w, http.StatusUnauthorized, coreerrors.ErrMissingToken)
		return
	}

	rate, err := strconv.ParseFloat(req.RateStr, 64)
	if err != nil {
		corejson.WriteError(w, http.StatusBadRequest, coreerrors.ErrInvalidRequest)
		return
	}

	output, err := h.u.CreateExchangeRate(r.Context(), &CreateExchangeRateInput{UserID: userID, TargetCurrency: req.TargetCurrency, Rate: rate})
	if err != nil {
		log.Println(err)
		switch err {
		case coreerrors.ErrInvalidRate:
			corejson.WriteError(w, http.StatusBadRequest, err)
		default:
			corejson.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
		}
		return
	}

	corejson.Write(w, http.StatusCreated, output)
}

func (h *ExchangeRateHandler) ListAllExchangeRates(w http.ResponseWriter, r *http.Request) {
	ers, err := h.u.ListAllExchangeRates(r.Context())
	if err != nil {
		log.Println(err)
		corejson.WriteError(w, http.StatusInternalServerError, coreerrors.ErrUnknown)
		return
	}

	if ers == nil {
		ers = []ExchangeRateItem{}
	}

	corejson.Write(w, http.StatusOK, ers)
}
