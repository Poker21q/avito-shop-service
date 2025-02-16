package handler

import (
	"context"
	"merch/internal/web/v1/pkg/response"
	"net/http"

	"github.com/gorilla/mux"
)

type PurchaseService interface {
	BuyItem(ctx context.Context, userID string, item string) error
}

type PurchaseLogger interface {
	Info(msg string)
	Error(msg string)
}

type BuyItemHandler struct {
	Service PurchaseService
	Logger  PurchaseLogger
}

func NewBuyItemHandler(service PurchaseService, logger PurchaseLogger) *BuyItemHandler {
	return &BuyItemHandler{
		Service: service,
		Logger:  logger,
	}
}

func (h *BuyItemHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(string)
	if !ok || userID == "" {
		h.Logger.Error("error extracting user_id from context or user_id is empty")
		response.Error(w, http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	item, exists := vars["item"]
	if !exists || item == "" {
		h.Logger.Error("item not specified or empty")
		response.Error(w, http.StatusBadRequest)
		return
	}

	err := h.Service.BuyItem(r.Context(), userID, item)
	if err != nil {
		h.Logger.Error("error buying item: " + err.Error())
		response.WithDomainError(w, err)
		return
	}

	h.Logger.Info("item bought successfully: " + item)
	response.Success(w, http.StatusOK)
}
