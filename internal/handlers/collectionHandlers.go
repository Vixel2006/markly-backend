package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

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
		log.Error().Err(err).Msg("Invalid JSON input for AddCollection")
		utils.SendJSONError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	col.UserID = userID
	col.ID = primitive.NewObjectID()

	collection := h.db.Client().Database("markly").Collection("collections")

	if err := utils.CreateUniqueIndex(collection, bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}}, "Collection name"); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Warn().Err(err).Msg("Collection name already exists")
			utils.SendJSONError(w, err.Error(), http.StatusConflict)
		} else {
			log.Error().Err(err).Msg("Failed to create index for collection")
			utils.SendJSONError(w, "Failed to set up collection", http.StatusInternalServerError)
		}
		return
	}

	_, err = collection.InsertOne(context.Background(), col)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Str("collection_name", col.Name).Str("user_id", userID.Hex()).Msg("Collection name already exists for this user")
			utils.SendJSONError(w, "Collection name already exists for this user.", http.StatusConflict)
		} else {
			log.Error().Err(err).Str("collection_name", col.Name).Str("user_id", userID.Hex()).Msg("Failed to insert collection")
			utils.SendJSONError(w, "Failed to insert collection", http.StatusInternalServerError)
		}
		return
	}

	log.Info().Str("collection_id", col.ID.Hex()).Str("collection_name", col.Name).Msg("Collection added successfully")
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
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Database error fetching collections")
		utils.SendJSONError(w, "Database error fetching collections", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var results []models.Collection
	if err := cursor.All(context.Background(), &results); err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error decoding collection results")
		utils.SendJSONError(w, "Error decoding collection results", http.StatusInternalServerError)
		return
	}

	log.Info().Int("count", len(results)).Str("user_id", userID.Hex()).Msg("Collections retrieved successfully")
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
		log.Warn().Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection not found or unauthorized")
		utils.SendJSONError(w, "Collection not found or unauthorized", http.StatusNotFound)
		return
	} else if err != nil {
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Database error finding collection")
		utils.SendJSONError(w, "Database error finding collection", http.StatusInternalServerError)
		return
	}

	log.Info().Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection retrieved successfully")
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
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Database error deleting collection")
		utils.SendJSONError(w, "Database error deleting collection", http.StatusInternalServerError)
		return
	}
	if result.DeletedCount == 0 {
		log.Warn().Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection not found or unauthorized to delete")
		utils.SendJSONError(w, "Collection not found or unauthorized", http.StatusNotFound)
		return
	}

	log.Info().Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection deleted successfully")
	utils.RespondWithJSON(w, http.StatusOK, bson.M{"message": "Collection deleted successfully", "deleted_count": result.DeletedCount})
}

func (h *CollectionHandler) buildCollectionUpdateFields(updatePayload models.CollectionUpdate) (bson.M, error) {
	updateFields := bson.M{}
	if updatePayload.Name != nil {
		updateFields["name"] = *updatePayload.Name
	}
	return updateFields, nil
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
		log.Error().Err(err).Msg("Invalid JSON payload for UpdateCollection")
		utils.SendJSONError(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updateFields, err := h.buildCollectionUpdateFields(updatePayload)
	if err != nil {
		log.Error().Err(err).Msg("Error building update fields for collection")
		utils.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(updateFields) == 0 {
		log.Warn().Msg("No fields to update for collection")
		utils.SendJSONError(w, "No fields to update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": collectionID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := h.db.Client().Database("markly").Collection("collections")

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection name already exists for this user")
			utils.SendJSONError(w, "Collection name already exists for this user.", http.StatusConflict)
			return
		}
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to update collection")
		utils.SendJSONError(w, "Failed to update collection", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		log.Warn().Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection not found or unauthorized to update")
		utils.SendJSONError(w, "Collection not found or unauthorized to update", http.StatusNotFound)
		return
	}

	var updatedCollection models.Collection
	err = collection.FindOne(context.Background(), filter).Decode(&updatedCollection)
	if err != nil {
		log.Error().Err(err).Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to find updated collection")
		utils.SendJSONError(w, "Failed to retrieve the updated collection", http.StatusInternalServerError)
		return
	}

	log.Info().Str("collection_id", collectionID.Hex()).Str("user_id", userID.Hex()).Msg("Collection updated successfully")
	utils.RespondWithJSON(w, http.StatusOK, updatedCollection)
}
