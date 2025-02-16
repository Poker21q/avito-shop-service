package middleware

import (
	"github.com/golang-jwt/jwt/v4"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var validToken string
var invalidTokenWithoutUserID string

func init() {
	claims := jwt.MapClaims{
		"user_id": "user123",
		"exp":     time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	var err error
	validToken, err = token.SignedString([]byte("testSecret"))
	if err != nil {
		panic("error generating valid token: " + err.Error())
	}

	claimsWithoutUserID := jwt.MapClaims{
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	tokenWithoutUserID := jwt.NewWithClaims(jwt.SigningMethodHS256, claimsWithoutUserID)
	invalidTokenWithoutUserID, err = tokenWithoutUserID.SignedString([]byte("testSecret"))
	if err != nil {
		panic("error generating invalid token without user_id: " + err.Error())
	}
}

type MockJWTLogger struct {
	mock.Mock
}

func (m *MockJWTLogger) Info(msg string) {}

func (m *MockJWTLogger) Error(msg string) {}

func TestJWT_Authenticate(t *testing.T) {
	tests := []struct {
		name         string
		authHeader   string
		expectedCode int
	}{
		{
			name:         "valid token",
			authHeader:   "Bearer " + validToken,
			expectedCode: http.StatusOK,
		},
		{
			name:         "missing authorization header",
			authHeader:   "",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "invalid token",
			authHeader:   "Bearer invalidToken",
			expectedCode: http.StatusUnauthorized,
		},
		{
			name:         "missing user_id in claims",
			authHeader:   "Bearer " + invalidTokenWithoutUserID,
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := new(MockJWTLogger)

			jwtMiddleware := NewJWT("testSecret", logger)

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req, _ := http.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.authHeader)
			resp := httptest.NewRecorder()

			jwtMiddleware.Authenticate(next).ServeHTTP(resp, req)

			assert.Equal(t, tt.expectedCode, resp.Code)
		})
	}
}
