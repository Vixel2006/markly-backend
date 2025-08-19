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

type CollectionHandler struct {
	db database.Service
}

func NewCollectionHandler(db database.Service) *CollectionHandler {
	return &CollectionHandler{db: db}
}

func (h *CollectionHandler) AddCollection(w http.ResponseWriter, r *http.Request) {
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

	var col models.Collection
	if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
		http.Error(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	col.UserID = userID
	col.ID = primitive.NewObjectID()

	collection := h.db.Client().Database("markly").Collection("collections")

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"name": 1, "user_id": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Printf("Duplicate collection name: %v", err)
			http.Error(w, "Collection name already exists for this user.", http.StatusConflict)
		} else {
			log.Printf("Failed to create index for collections: %v", err)
			http.Error(w, "Failed to set up collection", http.StatusInternalServerError)
		}
		return
	}

	_, err = collection.InsertOne(context.Background(), col)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Println("Collection name already exists for this user.")
			http.Error(w, "Collection name already exists for this user.", http.StatusConflict)
		} else {
			log.Printf("Failed to insert collection: %v", err)
			http.Error(w, "Failed to insert collection", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(col)
}

func (h *CollectionHandler) GetCollections(w http.ResponseWriter, r *http.Request) {
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

	collection := h.db.Client().Database("markly").Collection("collections")

	cursor, err := collection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		log.Printf("Database error fetching collections: %v", err)
		http.Error(w, "Database error fetching collections", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var results []models.Collection
	if err := cursor.All(context.Background(), &results); err != nil {
		log.Printf("Error decoding collection results: %v", err)
		http.Error(w, "Error decoding collection results", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
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

	if idStr == "" {
		http.Error(w, "Missing collection ID parameter", http.StatusBadRequest)
		return
	}

	collectionID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid collection ID format", http.StatusBadRequest)
		return
	}

	collection := h.db.Client().Database("markly").Collection("collections")

	var col models.Collection
	filter := bson.M{"_id": collectionID, "user_id": userID}
	err = collection.FindOne(context.Background(), filter).Decode(&col)
	if err == mongo.ErrNoDocuments {
		http.Error(w, "Collection not found or unauthorized", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Database error finding collection: %v", err)
		http.Error(w, "Database error finding collection", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(col)
}

func (h *CollectionHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
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

	if idStr == "" {
		http.Error(w, "Missing collection ID parameter", http.StatusBadRequest)
		return
	}

	collectionID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid collection ID format", http.StatusBadRequest)
		return
	}

	collection := h.db.Client().Database("markly").Collection("collections")

	filter := bson.M{"_id": collectionID, "user_id": userID}
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Database error deleting collection: %v", err)
		http.Error(w, "Database error deleting collection", http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		http.Error(w, "Collection not found or unauthorized", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"message": "Collection deleted successfully", "deleted_count": result.DeletedCount})
}

func (h *CollectionHandler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
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

	collectionID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid collection ID format", http.StatusBadRequest)
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

	filter := bson.M{"_id": collectionID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := h.db.Client().Database("markly").Collection("collections")

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Failed to update collection with ID %s for user %s: %v", idStr, userIDStr, err)
		http.Error(w, "Failed to update collection", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Collection not found or unauthorized to update", http.StatusNotFound)
		return
	}

	var updatedCollection models.Collection
	err = collection.FindOne(context.Background(), filter).Decode(&updatedCollection)
	if err != nil {
		log.Printf("Failed to find updated collection with ID %s for user %s: %v", idStr, userIDStr, err)
		http.Error(w, "Failed to retrieve the updated collection", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCollection)
}
