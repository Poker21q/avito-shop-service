package jwtutils

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name            string
		args            map[string]interface{}
		expiration      time.Duration
		secret          string
		expectedError   error
		expectedSuccess bool
	}{
		{
			name: "valid token generation",
			args: map[string]interface{}{
				"user_id": "12345",
			},
			expiration:      time.Hour,
			secret:          "mySecret",
			expectedError:   nil,
			expectedSuccess: true,
		},
		{
			name: "missing secret",
			args: map[string]interface{}{
				"user_id": "12345",
			},
			expiration:      time.Hour,
			secret:          "",
			expectedError:   jwt.ErrInvalidKey,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := Generate(tt.args, tt.expiration, tt.secret)

			if tt.expectedError != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tt.expectedError))
				assert.Empty(t, token)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)

				parsedToken, parseErr := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, ErrInvalidToken
					}
					return []byte(tt.secret), nil
				})

				assert.NoError(t, parseErr)

				claims, ok := parsedToken.Claims.(jwt.MapClaims)
				require.True(t, ok)

				assert.Equal(t, "12345", claims["user_id"])

				expTime := claims["exp"].(float64)
				assert.Greater(t, expTime, float64(time.Now().Unix()))
			}
		})
	}
}
