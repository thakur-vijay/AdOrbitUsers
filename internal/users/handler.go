package users

import (
	"net/http"

	"github.com/thakur-vijay/adorbit-go/apperrors"
	"github.com/thakur-vijay/adorbit-go/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var signupRequest SignupRequestDTO
	if err := response.DecodeJSONAndValidate(r, &signupRequest); err != nil {
		appError := apperrors.NewBadRequest(err.Error(), nil)
		response.Write(w, appError.StatusCode, appError.Message)
		return
	}
}

func (h *Handler) Fetch(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) FetchByID(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {

}
