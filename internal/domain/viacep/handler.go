package viacep

import (
	"log"
	"net/http"

	errors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/json"
	infraviacep "github.com/RuanHOliveira/estatehub_api/internal/infra/viacep"
	"github.com/go-chi/chi/v5"
)

type AddressResponse struct {
	ZipCode      string `json:"zip_code"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	Complement   string `json:"complement"`
}

type ViaCEPHandler struct {
	client infraviacep.ViaCEPClient
}

func NewViaCEPHandler(client infraviacep.ViaCEPClient) *ViaCEPHandler {
	return &ViaCEPHandler{client: client}
}

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
