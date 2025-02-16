package handler

import (
	"context"
	"encoding/json"
	"errors"
	"merch/internal/domain"
	"merch/internal/web/v1/dto"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockInfoService struct {
	mock.Mock
}

func (m *MockInfoService) GetUserInfo(ctx context.Context, userID string) (*domain.UserInfo, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.UserInfo), args.Error(1)
}

type MockInfoLogger struct {
	mock.Mock
}

func (m *MockInfoLogger) Info(msg string) {}

func (m *MockInfoLogger) Error(msg string) {}

func TestInfoHandler_Handle(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		setupMocks   func(service *MockInfoService)
		expectedCode int
		expectedErr  error
		expectedResp dto.InfoResponse
	}{
		{
			name:   "successful user info retrieval",
			userID: "user123",
			setupMocks: func(service *MockInfoService) {
				service.On("GetUserInfo", mock.Anything, "user123").Return(&domain.UserInfo{
					CoinBalance: 500,
					Inventory: []domain.UserInventory{
						{UserID: "user123", MerchName: "t-shirt", Quantity: 2},
						{UserID: "user123", MerchName: "mug", Quantity: 5},
					},
					CoinHistoryReceived: []domain.CoinTransfer{
						{FromUserID: "user456", Amount: 100, TransactionType: "received"},
					},
					CoinHistorySent: []domain.CoinTransfer{
						{ToUserID: "user789", Amount: 50, TransactionType: "sent"},
					},
				}, nil)
			},
			expectedCode: http.StatusOK,
			expectedErr:  nil,
			expectedResp: dto.InfoResponse{
				Coins: 500,
				Inventory: []dto.InfoResponseInventory{
					{Type_: "t-shirt", Quantity: 2},
					{Type_: "mug", Quantity: 5},
				},
				CoinHistory: &dto.InfoResponseCoinHistory{
					Received: []dto.InfoResponseCoinHistoryReceived{
						{FromUser: "user456", Amount: 100},
					},
					Sent: []dto.InfoResponseCoinHistorySent{
						{ToUser: "user789", Amount: 50},
					},
				},
			},
		},
		{
			name:         "missing user ID",
			userID:       "",
			expectedCode: http.StatusUnauthorized,
			expectedErr:  nil,
		},
		{
			name:   "user info retrieval failure",
			userID: "user123",
			setupMocks: func(service *MockInfoService) {
				service.On("GetUserInfo", mock.Anything, "user123").Return(new(domain.UserInfo), domain.ErrInternalServerError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  domain.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создание моков
			service := new(MockInfoService)
			logger := new(MockInfoLogger)
			logger.On("Error", mock.Anything).Return()
			logger.On("Info", mock.Anything).Return()

			handler := NewInfoHandler(service, logger)

			if tt.setupMocks != nil {
				tt.setupMocks(service)
			}

			req, _ := http.NewRequest(http.MethodGet, "/user-info", nil)
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
			req = req.WithContext(ctx)

			resp := httptest.NewRecorder()
			handler.Handle(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)

			if tt.expectedErr != nil {
				assert.True(t, errors.Is(tt.expectedErr, domain.ErrInternalServerError))
			}

			if tt.expectedCode == http.StatusOK {
				var actualResp dto.InfoResponse
				err := json.NewDecoder(resp.Body).Decode(&actualResp)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResp, actualResp)
			}

			if tt.name == "missing user ID" {
				service.AssertNotCalled(t, "GetUserInfo")
			} else {
				service.AssertExpectations(t)
			}
		})
	}
}
