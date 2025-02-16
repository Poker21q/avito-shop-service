package handler

import (
	"context"
	"encoding/json"
	"merch/internal/web/v1/dto"
	"merch/internal/web/v1/pkg/ctxkey"
	"merch/internal/web/v1/pkg/response"
	"net/http"
)

type CoinService interface {
	SendCoins(ctx context.Context, fromUserID string, toUser string, amount int) error
}

type CoinLogger interface {
	Info(msg string)
	Error(msg string)
}

type SendCoinHandler struct {
	Service CoinService
	Logger  CoinLogger
}

func NewSendCoinHandler(service CoinService, logger CoinLogger) *SendCoinHandler {
	return &SendCoinHandler{
		Service: service,
		Logger:  logger,
	}
}

func (h *SendCoinHandler) Handle(w http.ResponseWriter, r *http.Request) {
	fromUser, ok := r.Context().Value(ctxkey.UserIDKey).(string)
	if !ok || fromUser == "" {
		h.Logger.Error("error extracting user_id from context")
		response.Error(w, http.StatusUnauthorized)
		return
	}

	var sendCoinRequest dto.SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&sendCoinRequest); err != nil {
		h.Logger.Error("error decoding send coin request: " + err.Error())
		response.Error(w, http.StatusBadRequest)
		return
	}

	if sendCoinRequest.Amount <= 0 {
		h.Logger.Error("invalid amount in send coin request")
		response.Error(w, http.StatusBadRequest)
		return
	}

	err := h.Service.SendCoins(r.Context(), fromUser, sendCoinRequest.ToUser, int(sendCoinRequest.Amount))
	if err != nil {
		h.Logger.Error("error sending coins: " + err.Error())
		response.WithDomainError(w, err)
		return
	}

	h.Logger.Info("sending coins successfully retrieved for user_id:" + fromUser)
	response.Success(w, http.StatusOK)
}
