package middlewares

import (
	"context"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"markly/internal/utils"
	"net/http"
	"os"
	"strings"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwtKey := []byte(os.Getenv("JWT_SECRET"))
		if len(jwtKey) == 0 {
			log.Println("JWT_SECRET is not set in environment. Authentication will fail.")
			http.Error(w, "Server configuration error: JWT secret missing", http.StatusInternalServerError)
			return
		}
		log.Printf("jwtMiddleware: jwtKey length: %d", len(jwtKey))
		tokenString := r.Header.Get("Authorization")

		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		// Extract the token from the "Bearer <token>" format
		if !strings.HasPrefix(tokenString, "Bearer ") {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}
		tokenString = tokenString[len("Bearer "):]

		claims := &utils.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "userID", claims.ID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
