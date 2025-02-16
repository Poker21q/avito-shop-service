package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRecoveryLogger struct {
	mock.Mock
}

func (m *MockRecoveryLogger) Error(msg string) {
	m.Called(msg)
}

func TestRecoveryMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		handler      http.Handler
		triggerPanic bool
		expectedCode int
		expectedLog  string
	}{
		{
			name: "no panic",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}),
			triggerPanic: false,
			expectedCode: http.StatusOK,
			expectedLog:  "",
		},
		{
			name: "panic triggered",
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("something went wrong")
			}),
			triggerPanic: true,
			expectedCode: http.StatusInternalServerError,
			expectedLog:  "recovered from panic: something went wrong\nstack trace:\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := new(MockRecoveryLogger)

			if tt.triggerPanic {
				mockLogger.On("Error", mock.MatchedBy(func(msg string) bool {
					return len(msg) > len(tt.expectedLog) && msg[:len(tt.expectedLog)] == tt.expectedLog
				})).Return()
			}

			recoveryMiddleware := Recovery(mockLogger)
			handler := recoveryMiddleware(tt.handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			resp := httptest.NewRecorder()

			handler.ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)

			if tt.triggerPanic {
				mockLogger.AssertExpectations(t)
			} else {
				mockLogger.AssertNotCalled(t, "Error", mock.Anything)
			}
		})
	}
}
