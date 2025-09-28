package utils

import (
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"os"
	"time"
)

type Claims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

// Generate JWT
func GenerateJWT(id primitive.ObjectID) (string, error) {
	jwtKey := []byte(os.Getenv("JWT_SECRET"))
	log.Printf("jwtUtils: jwtKey length: %d", len(jwtKey))

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		ID: id.Hex(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}
