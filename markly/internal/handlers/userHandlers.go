package handlers

import (
	"encoding/json"
	"net/http"

	_ "github.com/joho/godotenv/autoload"

	"markly/internal/database"
)

type UserHandler struct {
	db *database.Service
}

func NewUserHandler(db *database.Service) *UserHandler {
	return &UserHandler{db: db}
}

func (u *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
  json.NewEncoder(w).Encode(map[string]string{"message": "Register new User"})
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Login with your email and password"})
}

func (u *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Signout of your account"})
}

