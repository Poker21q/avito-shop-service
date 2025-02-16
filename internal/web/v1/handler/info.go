package handler

import (
	"context"
	"merch/internal/domain"
	"merch/internal/web/v1/dto"
	"merch/internal/web/v1/pkg/ctxkey"
	"merch/internal/web/v1/pkg/response"
	"net/http"
)

type InfoService interface {
	GetUserInfo(ctx context.Context, userID string) (*domain.UserInfo, error)
}

type InfoLogger interface {
	Info(msg string)
	Error(msg string)
}

type InfoHandler struct {
	Service InfoService
	Logger  InfoLogger
}

func NewInfoHandler(service InfoService, logger InfoLogger) *InfoHandler {
	return &InfoHandler{
		Service: service,
		Logger:  logger,
	}
}

func (h *InfoHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkey.UserIDKey).(string)
	if !ok || userID == "" {
		h.Logger.Error("error extracting user_id from context")
		response.Error(w, http.StatusUnauthorized)
		return
	}

	userInfo, err := h.Service.GetUserInfo(r.Context(), userID)
	if err != nil {
		h.Logger.Error("error retrieving user info: " + err.Error())
		response.WithDomainError(w, err)
		return
	}

	infoResponse := mapToInfoResponse(userInfo)

	h.Logger.Info("user info successfully retrieved for user_id: " + userID)
	response.SuccessJSON(w, infoResponse, http.StatusOK)
}

func mapToInfoResponse(userInfo *domain.UserInfo) dto.InfoResponse {
	infoResponse := dto.InfoResponse{
		Coins: int32(userInfo.CoinBalance),
	}

	var inventory []dto.InfoResponseInventory
	for _, item := range userInfo.Inventory {
		inventory = append(inventory, dto.InfoResponseInventory{
			Type_:    item.MerchName,
			Quantity: int32(item.Quantity),
		})
	}
	infoResponse.Inventory = inventory

	infoResponse.CoinHistory = &dto.InfoResponseCoinHistory{
		Received: mapReceivedCoinHistory(userInfo.CoinHistoryReceived),
		Sent:     mapSentCoinHistory(userInfo.CoinHistorySent),
	}

	return infoResponse
}

func mapReceivedCoinHistory(received []domain.CoinTransfer) []dto.InfoResponseCoinHistoryReceived {
	var result []dto.InfoResponseCoinHistoryReceived
	for _, transfer := range received {
		result = append(result, dto.InfoResponseCoinHistoryReceived{
			FromUser: transfer.FromUserID,
			Amount:   int32(transfer.Amount),
		})
	}
	return result
}

func mapSentCoinHistory(sent []domain.CoinTransfer) []dto.InfoResponseCoinHistorySent {
	var result []dto.InfoResponseCoinHistorySent
	for _, transfer := range sent {
		result = append(result, dto.InfoResponseCoinHistorySent{
			ToUser: transfer.ToUserID,
			Amount: int32(transfer.Amount),
		})
	}
	return result
}
