package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils"
)

type CollectionHandler struct {
	db database.Service
}

func NewCollectionHandler(db database.Service) *CollectionHandler {
	return &CollectionHandler{db: db}
}

func (h *CollectionHandler) AddCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var col models.Collection
	if err := json.NewDecoder(r.Body).Decode(&col); err != nil {
		utils.SendJSONError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	col.UserID = userID
	col.ID = primitive.NewObjectID()

	collection := h.db.Client().Database("markly").Collection("collections")

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	_, err = collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Printf("Duplicate collection name: %v", err)
			utils.SendJSONError(w, "Collection name already exists for this user.", http.StatusConflict)
		} else {
			log.Printf("Failed to create index for collections: %v", err)
			utils.SendJSONError(w, "Failed to set up collection", http.StatusInternalServerError)
		}
		return
	}

	_, err = collection.InsertOne(context.Background(), col)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Println("Collection name already exists for this user.")
			utils.SendJSONError(w, "Collection name already exists for this user.", http.StatusConflict)
		} else {
			log.Printf("Failed to insert collection: %v", err)
			utils.SendJSONError(w, "Failed to insert collection", http.StatusInternalServerError)
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, col)
}

func (h *CollectionHandler) GetCollections(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	collection := h.db.Client().Database("markly").Collection("collections")

	cursor, err := collection.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		log.Printf("Database error fetching collections: %v", err)
		utils.SendJSONError(w, "Database error fetching collections", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var results []models.Collection
	if err := cursor.All(context.Background(), &results); err != nil {
		log.Printf("Error decoding collection results: %v", err)
		utils.SendJSONError(w, "Error decoding collection results", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, results)
}

func (h *CollectionHandler) GetCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	collectionID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	collection := h.db.Client().Database("markly").Collection("collections")

	var col models.Collection
	filter := bson.M{"_id": collectionID, "user_id": userID}
	err = collection.FindOne(context.Background(), filter).Decode(&col)
	if err == mongo.ErrNoDocuments {
		utils.SendJSONError(w, "Collection not found or unauthorized", http.StatusNotFound)
		return
	} else if err != nil {
		log.Printf("Database error finding collection: %v", err)
		utils.SendJSONError(w, "Database error finding collection", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, col)
}

func (h *CollectionHandler) DeleteCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	collectionID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	collection := h.db.Client().Database("markly").Collection("collections")

	filter := bson.M{"_id": collectionID, "user_id": userID}
	result, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Database error deleting collection: %v", err)
		utils.SendJSONError(w, "Database error deleting collection", http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		utils.SendJSONError(w, "Collection not found or unauthorized", http.StatusNotFound)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, bson.M{"message": "Collection deleted successfully", "deleted_count": result.DeletedCount})
}

func (h *CollectionHandler) UpdateCollection(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	collectionID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	var updatePayload models.CollectionUpdate
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		utils.SendJSONError(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updateFields := bson.M{}
	if updatePayload.Name != nil {
		updateFields["name"] = *updatePayload.Name
	}

	if len(updateFields) == 0 {
		utils.SendJSONError(w, "No fields to update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": collectionID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := h.db.Client().Database("markly").Collection("collections")

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			utils.SendJSONError(w, "Collection name already exists for this user.", http.StatusConflict)
			return
		}
		log.Printf("Failed to update collection with ID %s for user %s: %v", collectionID.Hex(), userID.Hex(), err)
		utils.SendJSONError(w, "Failed to update collection", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendJSONError(w, "Collection not found or unauthorized to update", http.StatusNotFound)
		return
	}

	var updatedCollection models.Collection
	err = collection.FindOne(context.Background(), filter).Decode(&updatedCollection)
	if err != nil {
		log.Printf("Failed to find updated collection with ID %s for user %s: %v", collectionID.Hex(), userID.Hex(), err)
		utils.SendJSONError(w, "Failed to retrieve the updated collection", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, updatedCollection)
}
