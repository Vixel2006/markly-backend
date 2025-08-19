package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

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
		http.Error(w, "Invalid user ID.", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format.", http.StatusUnauthorized)
		return
	}

	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	category.ID = primitive.NewObjectID()
	category.UserID = userID // Ensure UserID is set from context

	collection := h.db.Client().Database("markly").Collection("categories")

	// --- FIX IS HERE: Use bson.D for ordered keys in compound index ---
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}}, // Corrected to bson.D
		Options: options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			http.Error(w, "Category name already exists for this user.", http.StatusConflict)
			log.Printf("Duplicate category name: %v", err)
			return
		}
		// This log will now show the actual error if it's not a duplicate key error
		log.Printf("Failed to create index: %v", err)
		http.Error(w, "Failed to set up category collection", http.StatusInternalServerError)
		return
	}

	_, err = collection.InsertOne(context.Background(), category)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Println("Category name already exists for this user.")
			http.Error(w, "Category name already exists for this user.", http.StatusConflict)
		} else {
			log.Printf("Failed to insert category: %v", err)
			http.Error(w, "Failed to insert category", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(category)
}

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	var categories []models.Category

	collection := h.db.Client().Database("markly").Collection("categories")

	filter := bson.M{"user_id": userID}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error finding categories: %v", err)
		http.Error(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &categories); err != nil {
		log.Printf("Error decoding categories: %v", err)
		http.Error(w, "Error decoding categories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	categoryID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID format", http.StatusBadRequest)
		return
	}

	collection := h.db.Client().Database("markly").Collection("categories")

	filter := bson.M{"user_id": userID, "_id": categoryID}

	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Failed to delete category with ID %s for user %s: %v", idStr, userIDStr, err)
		http.Error(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		http.Error(w, "Category not found or unauthorized", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"message": "Category deleted successfully", "deleted_count": deleteResult.DeletedCount})
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]

	categoryID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID format", http.StatusBadRequest)
		return
	}

	var updatePayload struct {
		Name *string `json:"name,omitempty" bson:"name,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updateFields := bson.M{}
	if updatePayload.Name != nil {
		updateFields["name"] = *updatePayload.Name
	}

	if len(updateFields) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": categoryID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := h.db.Client().Database("markly").Collection("categories")

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Failed to update category with ID %s for user %s: %v", idStr, userIDStr, err)
		http.Error(w, "Failed to update category", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Category not found or unauthorized to update", http.StatusNotFound)
		return
	}

	var updatedCategory models.Category
	err = collection.FindOne(context.Background(), filter).Decode(&updatedCategory)
	if err != nil {
		log.Printf("Failed to find updated category with ID %s for user %s: %v", idStr, userIDStr, err)
		http.Error(w, "Failed to retrieve the updated category", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCategory)
}
