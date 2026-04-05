package handler

import (
	"net/http"
	"strings"
)

// BearerAuth returns a middleware that requires a valid Authorization: Bearer <token> header.
// Requests with a missing or incorrect token are rejected with 401.
func BearerAuth(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			val := r.Header.Get("Authorization")
			if !strings.HasPrefix(val, "Bearer ") || strings.TrimPrefix(val, "Bearer ") != token {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
