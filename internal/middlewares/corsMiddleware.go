package middlewares

import (
	"net/http"
	"os"
	"strings"
)

var allowedOrigins []string

func init() {
	allowed := os.Getenv("ALLOWED_ORIGINS")
	allowedOrigins = strings.Split(allowed, ",")
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		for _, allowed := range allowedOrigins {
			if strings.TrimSpace(allowed) == origin {
				w.Header().Set("Access-Control-Allow-Origin", allowed)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				break
			}
		}

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
