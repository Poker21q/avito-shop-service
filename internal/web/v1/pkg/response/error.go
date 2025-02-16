package response

import (
	"encoding/json"
	"errors"
	"merch/internal/domain"
	"merch/internal/web/v1/dto"
	"net/http"
)

func Error(w http.ResponseWriter, statusCode int) {
	var body dto.ErrorResponse

	switch statusCode {
	case http.StatusBadRequest:
		body = dto.ErrorResponse{Errors: "bad request"}
	case http.StatusUnauthorized:
		body = dto.ErrorResponse{Errors: "unauthorized"}
	default:
		body = dto.ErrorResponse{Errors: "internal server error"}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

func WithDomainError(w http.ResponseWriter, err error) {
	var statusCode int

	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		statusCode = http.StatusUnauthorized
	case errors.Is(err, domain.ErrInternalServerError):
		statusCode = http.StatusInternalServerError
	default:
		statusCode = http.StatusBadRequest
	}

	Error(w, statusCode)
}
