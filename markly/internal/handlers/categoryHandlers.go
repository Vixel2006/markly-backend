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

func (h *CategoryHandler) AddCategory(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)

	if !ok {
		http.Error(w, "Invalid ID.", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)

	if err != nil {
		http.Error(w, "Invalid user ID format.", http.StatusUnauthorized)
		return
	}

	var category Category

	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	collection := h.db.Client().Database("makrly").Collection("categories")

	indexModel := mongo.IndexModel{
		Keys: bson.M{"Name": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context, indexModel)

	if err != nil {
		log.Fatal(err)
	}

	_, err = collection.InsertOne(context.TODO(), category)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Println("Already Exists")
		} else {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)

	if !ok {
		http.Error(w, "Invalid ID", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)

	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusUnauthorized)
	}

	var categories []Category

	collection := h.db.Client().Database("markly").collection("categories")

	filter := bson.M{"user_id": userID}

	cursor, err := collection.Find(context.Background(), filter)

	if err != nil {
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}

	if err := cursor.All(context.Background(), &categories); err != nil {
		http.Error(w, "Error decoding categories", http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userIdStr, ok := r.Context().Value("userID").(string)

	if !ok {
		http.Error(w, "Invalid ID", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)

	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusUnauthorized)
	}

	vars := mux.Vars(r)
	id := vars["id"]

	collection := h.db.Client().Database("markly").Collection("categories")

	filter := bson.M{"user_id", userID, "_id": id}

	delete_result, err := collection.DeleteOne(context.Background(), filter)


	if err != nil {
		http.Error(w, "Cannot delete the bookmark", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(delete_result)
}

