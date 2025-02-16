package handler

import (
	"bytes"
	"context"
	"errors"
	"merch/internal/domain"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCoinService struct {
	mock.Mock
}

func (m *MockCoinService) SendCoins(ctx context.Context, fromUserID string, toUser string, amount int) error {
	args := m.Called(ctx, fromUserID, toUser, amount)
	return args.Error(0)
}

type MockCoinLogger struct {
	mock.Mock
}

func (m *MockCoinLogger) Info(msg string) {}

func (m *MockCoinLogger) Error(msg string) {}

func TestSendCoinHandler_Handle(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		sendCoinReq  string
		setupMocks   func(service *MockCoinService)
		expectedCode int
		expectedErr  error
	}{
		{
			name:        "successful coin transfer",
			userID:      "user1",
			sendCoinReq: `{"toUser": "user2","amount": 100}`,
			setupMocks: func(service *MockCoinService) {
				service.On("SendCoins", mock.Anything, "user1", "user2", 100).Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedErr:  nil,
		},
		{
			name:         "missing user ID",
			userID:       "",
			sendCoinReq:  `{"toUser": "user2", "amount": 100}`,
			expectedCode: http.StatusUnauthorized,
			expectedErr:  nil,
		},
		{
			name:         "invalid amount",
			userID:       "user1",
			sendCoinReq:  `{"toUser": "user2", "amount": -10}`,
			expectedCode: http.StatusBadRequest,
			expectedErr:  nil,
		},
		{
			name:        "send coins failure - insufficient funds",
			userID:      "user1",
			sendCoinReq: `{"toUser": "user2", "amount": 100}`,
			setupMocks: func(service *MockCoinService) {
				service.On("SendCoins", mock.Anything, "user1", "user2", 100).Return(domain.ErrInsufficientFunds)
			},
			expectedCode: http.StatusBadRequest,
			expectedErr:  domain.ErrInsufficientFunds,
		},
		{
			name:        "send coins failure - internal error",
			userID:      "user1",
			sendCoinReq: `{"toUser": "user2","amount": 100}`,
			setupMocks: func(service *MockCoinService) {
				service.On("SendCoins", mock.Anything, "user1", "user2", 100).Return(domain.ErrInternalServerError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  domain.ErrInternalServerError,
		},
		{
			name:         "invalid request body format",
			userID:       "user1",
			sendCoinReq:  `{"toUser": "user2", "amount": "invalid"}`,
			expectedCode: http.StatusBadRequest,
			expectedErr:  nil,
		},
		{
			name:         "missing required fields in body",
			userID:       "user1",
			sendCoinReq:  `{"toUser": "user2"}`,
			expectedCode: http.StatusBadRequest,
			expectedErr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := new(MockCoinService)
			logger := new(MockCoinLogger)
			logger.On("Error", mock.Anything).Return()
			logger.On("Info", mock.Anything).Return()

			handler := NewSendCoinHandler(service, logger)

			if tt.setupMocks != nil {
				tt.setupMocks(service)
			}

			req, _ := http.NewRequest(http.MethodPost, "/sendCoin", bytes.NewReader([]byte(tt.sendCoinReq)))
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
			req = req.WithContext(ctx)

			resp := httptest.NewRecorder()
			handler.Handle(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)

			if tt.expectedErr != nil {
				assert.True(t, errors.Is(tt.expectedErr, domain.ErrInsufficientFunds) || errors.Is(tt.expectedErr, domain.ErrInternalServerError))
			}

			if tt.name == "missing user ID" || tt.name == "invalid amount" {
				service.AssertNotCalled(t, "SendCoins")
			} else {
				service.AssertExpectations(t)
			}
		})
	}
}
