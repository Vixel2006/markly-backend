package handlers

import (
	"strings"
	"log"
	"encoding/json"
	"net/http"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/gorilla/mux"
	"context"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"markly/internal/database"
	"markly/internal/models"
)

type CategoryHandler struct {
	db database.Service
}

func NewCategoryHandler(db database.Service) *CategoryHandler {
	return &CategoryHandler{db: db}
}

func (h *CategoryHandler) AddCategory(w http.ResponseWriter, r *http.Request) {}

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {}

func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {}

