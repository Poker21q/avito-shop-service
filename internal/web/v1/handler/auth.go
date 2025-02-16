package handler

import (
	"context"
	"encoding/json"
	"merch/internal/web/v1/dto"
	"merch/internal/web/v1/pkg/response"
	"net/http"
)

type AuthService interface {
	Auth(ctx context.Context, username, password string) (string, error)
}

type AuthLogger interface {
	Info(msg string)
	Error(msg string)
}

type AuthHandler struct {
	Service AuthService
	Logger  AuthLogger
}

func NewAuthHandler(service AuthService, logger AuthLogger) *AuthHandler {
	return &AuthHandler{
		Service: service,
		Logger:  logger,
	}
}

func (h *AuthHandler) Handle(w http.ResponseWriter, r *http.Request) {
	var authRequest dto.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&authRequest); err != nil {
		h.Logger.Error("error decoding auth request: " + err.Error())
		response.Error(w, http.StatusBadRequest)
		return
	}

	token, err := h.Service.Auth(r.Context(), authRequest.Username, authRequest.Password)
	if err != nil {
		h.Logger.Error("error during auth: " + err.Error())
		response.WithDomainError(w, err)
		return
	}

	h.Logger.Info("auth request finished successfully")
	response.SuccessJSON(w, dto.AuthResponse{Token: token}, http.StatusOK)
}
