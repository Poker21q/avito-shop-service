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

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Auth(ctx context.Context, username, password string) (string, error) {
	args := m.Called(ctx, username, password)
	return args.String(0), args.Error(1)
}

type MockAuthLogger struct {
	mock.Mock
}

func (m *MockAuthLogger) Info(msg string) {}

func (m *MockAuthLogger) Error(msg string) {}

func TestAuthHandler_Handle(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  string
		setupMocks   func(service *MockAuthService)
		expectedCode int
		expectedErr  error
	}{
		{
			name:        "successful auth",
			requestBody: `{"username":"testuser","password":"password"}`,
			setupMocks: func(service *MockAuthService) {
				service.On("Auth", mock.Anything, "testuser", "password").Return("valid_token", nil)
			},
			expectedCode: http.StatusOK,
			expectedErr:  nil,
		},
		{
			name:         "invalid JSON request",
			requestBody:  "invalid-json",
			expectedCode: http.StatusBadRequest,
			expectedErr:  nil,
		},
		{
			name:        "auth failure",
			requestBody: `{"username":"testuser","password":"wrongpassword"}`,
			setupMocks: func(service *MockAuthService) {
				service.On("Auth", mock.Anything, "testuser", "wrongpassword").Return("", domain.ErrInvalidCredentials)
			},
			expectedCode: http.StatusUnauthorized,
			expectedErr:  domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := new(MockAuthService)
			logger := new(MockAuthLogger)
			logger.On("Error", mock.Anything).Return()
			logger.On("Info", mock.Anything).Return()

			handler := NewAuthHandler(service, logger)

			if tt.setupMocks != nil {
				tt.setupMocks(service)
			}

			req, _ := http.NewRequest(http.MethodPost, "/auth", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()

			handler.Handle(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
			if tt.expectedErr != nil {
				assert.True(t, errors.Is(tt.expectedErr, domain.ErrInvalidCredentials) || errors.Is(tt.expectedErr, domain.ErrInternalServerError))
			}

			if tt.name == "invalid JSON request" {
				service.AssertNotCalled(t, "Auth")
			}

			service.AssertExpectations(t)
		})
	}
}
