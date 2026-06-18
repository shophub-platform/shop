package middleware

import (
	"net/http"

	"github.com/shophub/shop/pkg/response"
)

// InternalKeyAuth protects endpoints meant for internal services (e.g. the blockchain listener).
// Callers must supply the correct secret in the X-Internal-Key header.
func InternalKeyAuth(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Internal-Key") != key {
				response.Error(w, http.StatusUnauthorized, "invalid internal key")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
