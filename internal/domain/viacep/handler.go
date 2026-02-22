package viacep

import (
	"log"
	"net/http"

	errors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/json"
	infraviacep "github.com/RuanHOliveira/estatehub_api/internal/infra/viacep"
	"github.com/go-chi/chi/v5"
)

// AddressResponse representa o endereço retornado pela consulta de CEP.
type AddressResponse struct {
	ZipCode      string `json:"zip_code" example:"01310100"`
	Street       string `json:"street" example:"Avenida Paulista"`
	Neighborhood string `json:"neighborhood" example:"Bela Vista"`
	City         string `json:"city" example:"São Paulo"`
	State        string `json:"state" example:"SP"`
	Complement   string `json:"complement" example:""`
}

type ViaCEPHandler struct {
	client infraviacep.ViaCEPClient
}

func NewViaCEPHandler(client infraviacep.ViaCEPClient) *ViaCEPHandler {
	return &ViaCEPHandler{client: client}
}

// FindAddress godoc
// @Summary      Consultar endereço por CEP
// @Description  Busca informações de endereço via serviço ViaCEP. O CEP deve ter exatamente 8 dígitos numéricos, sem hífen.
// @Tags         viacep
// @Produce      json
// @Param        cep path string true "CEP (8 dígitos numéricos, sem hífen)" example("01310100")
// @Success      200 {object} AddressResponse
// @Failure      400 {object} json.ErrorResponse
// @Failure      404 {object} json.ErrorResponse
// @Failure      502 {object} json.ErrorResponse
// @Security     BearerAuth
// @Router       /viacep/{cep} [get]
func (h *ViaCEPHandler) FindAddress(w http.ResponseWriter, r *http.Request) {
	cep := chi.URLParam(r, "cep")

	address, err := h.client.FindAddressByCEP(r.Context(), cep)
	if err != nil {
		log.Println(err)

		switch err {
		case errors.ErrInvalidCEP:
			json.WriteError(w, http.StatusBadRequest, err)
		case errors.ErrCEPNotFound:
			json.WriteError(w, http.StatusNotFound, err)
		case errors.ErrExternalBadRequest:
			json.WriteError(w, http.StatusBadGateway, err)
		default:
			json.WriteError(w, http.StatusBadGateway, errors.ErrExternalServiceFailure)
		}
		return
	}

	json.Write(w, http.StatusOK, AddressResponse{
		ZipCode:      address.ZipCode,
		Street:       address.Street,
		Neighborhood: address.Neighborhood,
		City:         address.City,
		State:        address.State,
		Complement:   address.Complement,
	})
}
