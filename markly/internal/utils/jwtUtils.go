package utils

import (
	"log"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"os"
  "github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Claims struct {
	ID string `json:"id"`
	jwt.RegisteredClaims
}

// Generate JWT
func GenerateJWT(id primitive.ObjectID) (string, error) {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	jwtKey := []byte(os.Getenv("JWT_SECRET"))

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
