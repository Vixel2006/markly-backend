package middlewares

import (
	"context"
	"net/http"
	"github.com/golang-jwt/jwt/v5"
	"markly/internal/utils"
)

var jwtKey = []byte("super_secret_key")

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tokenString := r.Header.Get("Authorization")
        if tokenString == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }

        claims := &utils.Claims{}
        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return jwtKey, nil
        })

        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Attach claims (user info) to request context
        ctx := context.WithValue(r.Context(), "userEmail", claims.Email)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

