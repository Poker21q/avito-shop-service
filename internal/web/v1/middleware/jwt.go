package middleware

import (
	"context"
	"github.com/golang-jwt/jwt/v4"
	"merch/internal/web/v1/pkg/ctxkey"
	"merch/internal/web/v1/pkg/response"
	"net/http"
	"strings"
)

type JWTLogger interface {
	Info(msg string)
	Error(msg string)
}

type JWT struct {
	secret string
	logger JWTLogger
}

func NewJWT(secret string, logger JWTLogger) *JWT {
	return &JWT{secret: secret, logger: logger}
}

func (j *JWT) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			j.logger.Error("missing authorization header")
			response.Error(w, http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, &claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(j.secret), nil
		})

		if err != nil {
			j.logger.Error("invalid token: " + err.Error())
			response.Error(w, http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			j.logger.Error("token is not valid")
			response.Error(w, http.StatusUnauthorized)
			return
		}

		userID, ok := claims["user_id"].(string)
		if !ok {
			j.logger.Error("invalid token payload: missing user_id")
			response.Error(w, http.StatusUnauthorized)
			return
		}

		j.logger.Info("authentication successful for user: " + userID)
		ctx := context.WithValue(r.Context(), ctxkey.UserIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
