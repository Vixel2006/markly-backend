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

type CollectionHandler struct {
	db database.Service
}

func NewCollectionHandler(db database.Service) *CollectionHandler {
	return &CollectionHandler{db: db}
}

func (h *CollectionHandler) AddCollection(w http.ResponseWriter, r *http.Request) {}

func (h *CollectionHandler) GetCollections(w http.ResponseWriter, r *http.Request) {}

func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {}

func (h *CollectionHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {}

