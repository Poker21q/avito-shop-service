package middleware

import (
	"fmt"
	"merch/internal/web/v1/pkg/response"
	"net/http"
	"runtime/debug"
)

type RecoveryLogger interface {
	Error(msg string)
}

func Recovery(logger RecoveryLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error(fmt.Sprintf("recovered from panic: %v\nstack trace:\n%s", err, debug.Stack()))
					response.Error(w, http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
