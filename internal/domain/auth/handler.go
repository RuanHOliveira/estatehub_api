package auth

import (
	"log"
	"net/http"

	errors "github.com/RuanHOliveira/estatehub_api/internal/core/error"
	"github.com/RuanHOliveira/estatehub_api/internal/core/json"
)

type AuthHandler struct {
	u AuthUsecase
}

func NewAuthHandler(u AuthUsecase) *AuthHandler {
	return &AuthHandler{u: u}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.Read(r, &req); err != nil {
		log.Println(err)
		json.WriteError(w, http.StatusBadRequest, errors.ErrInvalidRequest)
		return
	}

	output, err := h.u.Register(r.Context(), &RegisterInput{Email: req.Email, Name: req.Name, Password: req.Password})
	if err != nil {
		log.Println(err)

		if err == errors.ErrEmailAlreadyUsed {
			json.WriteError(w, http.StatusConflict, err)
			return
		}

		json.WriteError(w, http.StatusInternalServerError, errors.ErrUnknown)
		return
	}

	json.Write(w, http.StatusCreated, RegisterResponse{User: output.User, AccessToken: output.AccessToken})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.Read(r, &req); err != nil {
		log.Println(err)
		json.WriteError(w, http.StatusBadRequest, errors.ErrInvalidRequest)
		return
	}

	output, err := h.u.Login(r.Context(), &LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		if err == errors.ErrUserNotFound {
			json.WriteError(w, http.StatusNotFound, err)
			return
		}

		if err == errors.ErrInvalidCredentials {
			json.WriteError(w, http.StatusUnauthorized, err)
			return
		}

		json.WriteError(w, http.StatusInternalServerError, errors.ErrUnknown)
		return
	}

	json.Write(w, http.StatusOK, LoginResponse{User: output.User, AccessToken: output.AccessToken})
}
