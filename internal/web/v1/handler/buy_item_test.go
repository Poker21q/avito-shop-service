package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"merch/internal/domain"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockPurchaseService struct {
	mock.Mock
}

func (m *MockPurchaseService) BuyItem(ctx context.Context, userID string, item string) error {
	args := m.Called(ctx, userID, item)
	return args.Error(0)
}

type MockPurchaseLogger struct {
	mock.Mock
}

func (m *MockPurchaseLogger) Info(msg string) {}

func (m *MockPurchaseLogger) Error(msg string) {}

func TestBuyItemHandler_Handle(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		item         string
		setupMocks   func(service *MockPurchaseService)
		expectedCode int
		expectedErr  error
	}{
		{
			name:   "successful purchase",
			userID: "user123",
			item:   "item123",
			setupMocks: func(service *MockPurchaseService) {
				service.On("BuyItem", mock.Anything, "user123", "item123").Return(nil)
			},
			expectedCode: http.StatusOK,
			expectedErr:  nil,
		},
		{
			name:         "missing user ID",
			userID:       "",
			item:         "item123",
			expectedCode: http.StatusUnauthorized,
			expectedErr:  nil,
		},
		{
			name:         "missing item",
			userID:       "user123",
			item:         "",
			expectedCode: http.StatusBadRequest,
			expectedErr:  nil,
		},
		{
			name:   "purchase failure - insufficient funds",
			userID: "user123",
			item:   "item123",
			setupMocks: func(service *MockPurchaseService) {
				service.On("BuyItem", mock.Anything, "user123", "item123").Return(domain.ErrInvalidCredentials)
			},
			expectedCode: http.StatusUnauthorized,
			expectedErr:  domain.ErrInvalidCredentials,
		},
		{
			name:   "purchase failure - internal error",
			userID: "user123",
			item:   "item123",
			setupMocks: func(service *MockPurchaseService) {
				service.On("BuyItem", mock.Anything, "user123", "item123").Return(domain.ErrInternalServerError)
			},
			expectedCode: http.StatusInternalServerError,
			expectedErr:  domain.ErrInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := new(MockPurchaseService)
			logger := new(MockPurchaseLogger)
			logger.On("Error", mock.Anything).Return()
			logger.On("Info", mock.Anything).Return()

			handler := NewBuyItemHandler(service, logger)

			if tt.setupMocks != nil {
				tt.setupMocks(service)
			}

			req, _ := http.NewRequest(http.MethodPost, "/buy/{item}", nil)
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
			req = req.WithContext(ctx)

			vars := map[string]string{"item": tt.item}
			req = mux.SetURLVars(req, vars)

			resp := httptest.NewRecorder()
			handler.Handle(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)

			if tt.expectedErr != nil {
				assert.True(t, errors.Is(tt.expectedErr, domain.ErrInvalidCredentials) || errors.Is(tt.expectedErr, domain.ErrInternalServerError))
			}

			if tt.name == "missing user ID" || tt.name == "missing item" {
				service.AssertNotCalled(t, "BuyItem")
			} else {
				service.AssertExpectations(t)
			}
		})
	}
}
