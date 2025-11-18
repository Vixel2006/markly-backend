package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils"
)

type TagHandler struct {
	db database.Service
}

func NewTagHandler(db database.Service) *TagHandler {
	return &TagHandler{db: db}
}

func (h *TagHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var tag models.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		log.Error().Err(err).Msg("Invalid JSON input for AddTag")
		utils.SendJSONError(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	tag.ID = primitive.NewObjectID()
	tag.UserID = userID
	tag.WeeklyCount = 0
	tag.PrevCount = 0
	tag.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

	collection := h.db.Client().Database("markly").Collection("tags")

	if err := utils.CreateUniqueIndex(collection, bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}}, "Tag name"); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Warn().Err(err).Msg("Tag name already exists")
			utils.SendJSONError(w, err.Error(), http.StatusConflict)
		} else {
			log.Error().Err(err).Msg("Failed to create index for tag")
			utils.SendJSONError(w, "Failed to set up tag collection", http.StatusInternalServerError)
		}
		return
	}

	_, err = collection.InsertOne(context.Background(), tag)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Str("tag_name", tag.Name).Str("user_id", userID.Hex()).Msg("Tag name already exists for this user")
			utils.SendJSONError(w, "Tag name already exists for this user.", http.StatusConflict)
		} else {
			log.Error().Err(err).Str("tag_name", tag.Name).Str("user_id", userID.Hex()).Msg("Failed to insert tag")
			utils.SendJSONError(w, "Failed to insert tag", http.StatusInternalServerError)
		}
		return
	}

	log.Info().Str("tag_id", tag.ID.Hex()).Str("tag_name", tag.Name).Msg("Tag added successfully")
	utils.RespondWithJSON(w, http.StatusCreated, tag)
}

func (h *TagHandler) fetchTagByID(ctx context.Context, userID, tagID primitive.ObjectID) (*models.Tag, error) {
	var tag models.Tag
	filter := bson.M{"_id": tagID, "user_id": userID}
	collection := h.db.Client().Database("markly").Collection("tags")
	err := collection.FindOne(ctx, filter).Decode(&tag)
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func (h *TagHandler) GetTagsByID(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	ctx := r.Context()
	ids := r.URL.Query()["id"]

	if len(ids) == 0 {
		log.Debug().Msg("No tag IDs provided, returning empty list")
		utils.RespondWithJSON(w, http.StatusOK, []models.Tag{})
		return
	}

	type result struct {
		Tag models.Tag
		Err error
	}

	resultsChan := make(chan result, len(ids))
	var wg sync.WaitGroup
	wg.Add(len(ids))

	for _, idStr := range ids {
		idStr := idStr
		go func() {
			defer wg.Done()

			objID, err := primitive.ObjectIDFromHex(strings.TrimSpace(idStr))
			if err != nil {
				log.Error().Err(err).Str("tag_id_string", idStr).Msg("Invalid tag ID format")
				resultsChan <- result{Err: err}
				return
			}

			tag, err := h.fetchTagByID(ctx, userID, objID)
			if err != nil {
				log.Error().Err(err).Str("tag_id", idStr).Str("user_id", userID.Hex()).Msg("Error finding tag")
				resultsChan <- result{Err: err}
				return
			}

			resultsChan <- result{Tag: *tag, Err: nil}
		}()
	}

	wg.Wait()
	close(resultsChan)

	var tags []models.Tag
	for r := range resultsChan {
		if r.Err == nil {
			tags = append(tags, r.Tag)
		}
	}

	log.Info().Int("count", len(tags)).Str("user_id", userID.Hex()).Msg("Tags retrieved by ID successfully")
	utils.RespondWithJSON(w, http.StatusOK, tags)
}

func (h *TagHandler) GetUserTags(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var tags []models.Tag
	collection := h.db.Client().Database("markly").Collection("tags")

	filter := bson.M{"user_id": userID}
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error finding tags for user")
		utils.SendJSONError(w, "Failed to retrieve tags", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &tags); err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error decoding tags")
		utils.SendJSONError(w, "Error decoding tags", http.StatusInternalServerError)
		return
	}

	log.Info().Int("count", len(tags)).Str("user_id", userID.Hex()).Msg("User tags retrieved successfully")
	utils.RespondWithJSON(w, http.StatusOK, tags)
}

func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	tagID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	collection := h.db.Client().Database("markly").Collection("tags")
	filter := bson.M{"_id": tagID, "user_id": userID}

	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to delete tag")
		utils.SendJSONError(w, "Failed to delete tag", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		log.Warn().Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Tag not found or unauthorized to delete")
		utils.SendJSONError(w, "Tag not found or unauthorized", http.StatusNotFound)
		return
	}

	log.Info().Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Tag deleted successfully")
	utils.RespondWithJSON(w, http.StatusOK, bson.M{"message": "Tag deleted successfully", "deleted_count": deleteResult.DeletedCount})
}

func (h *TagHandler) buildTagUpdateFields(updatePayload models.TagUpdate) (bson.M, error) {
	updateFields := bson.M{}
	if updatePayload.Name != nil {
		updateFields["name"] = *updatePayload.Name
	}
	if updatePayload.WeeklyCount != nil {
		updateFields["weekly_count"] = *updatePayload.WeeklyCount
	}
	if updatePayload.PrevCount != nil {
		updateFields["prev_count"] = *updatePayload.PrevCount
	}
	return updateFields, nil
}

func (h *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	tagID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	var updatePayload models.TagUpdate
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		log.Error().Err(err).Msg("Invalid JSON payload for UpdateTag")
		utils.SendJSONError(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updateFields, err := h.buildTagUpdateFields(updatePayload)
	if err != nil {
		log.Error().Err(err).Msg("Error building update fields for tag")
		utils.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(updateFields) == 0 {
		log.Warn().Msg("No fields to update for tag")
		utils.SendJSONError(w, "No fields to update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": tagID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := h.db.Client().Database("markly").Collection("tags")

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Tag name already exists for this user")
			utils.SendJSONError(w, "Tag name already exists for this user.", http.StatusConflict)
		} else {
			log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to update tag")
			utils.SendJSONError(w, "Failed to update tag", http.StatusInternalServerError)
		}
		return
	}

	if result.MatchedCount == 0 {
		log.Warn().Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Tag not found or unauthorized to update")
		utils.SendJSONError(w, "Tag not found or unauthorized", http.StatusNotFound)
		return
	}

	var updatedTag models.Tag
	err = collection.FindOne(context.Background(), filter).Decode(&updatedTag)
	if err != nil {
		log.Error().Err(err).Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to find updated tag")
		utils.SendJSONError(w, "Failed to retrieve the updated tag", http.StatusInternalServerError)
		return
	}

	log.Info().Str("tag_id", tagID.Hex()).Str("user_id", userID.Hex()).Msg("Tag updated successfully")
	utils.RespondWithJSON(w, http.StatusOK, updatedTag)
}
