package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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

type BookmarkHandler struct {
	db database.Service
}

func NewBookmarksHandler(db database.Service) *BookmarkHandler {
	return &BookmarkHandler{db: db}
}

func (h *BookmarkHandler) buildBookmarkFilter(r *http.Request, userID primitive.ObjectID) (bson.M, error) {
	filter := bson.M{"user_id": userID}

	tagsParam := r.URL.Query().Get("tags")
	if tagsParam != "" {
		tagsIDs, err := utils.ParseObjectIDs(tagsParam)
		if err != nil {
			return nil, fmt.Errorf("invalid tags ID format. Tags must be comma-separated hexadecimal ObjectIDs.")
		}
		filter["tagsid"] = bson.M{"$in": tagsIDs}
	}

	categoryParam := r.URL.Query().Get("category")
	if categoryParam != "" {
		categoryID, err := primitive.ObjectIDFromHex(categoryParam)
		if err != nil {
			return nil, fmt.Errorf("invalid category ID format. Category must be a hexadecimal ObjectID.")
		}
		filter["categoryid"] = categoryID
	}

	collectionsParam := r.URL.Query().Get("collections")
	if collectionsParam != "" {
		collectionIDs, err := utils.ParseObjectIDs(collectionsParam)
		if err != nil {
			return nil, fmt.Errorf("invalid collections ID format. Collections must be comma-separated hexadecimal ObjectIDs.")
		}
		filter["collectionsid"] = bson.M{"$in": collectionIDs}
	}

	isFavParam := r.URL.Query().Get("isFav")
	if isFavParam != "" {
		isFav, err := strconv.ParseBool(isFavParam)
		if err != nil {
			return nil, fmt.Errorf("invalid isFav format. Must be 'true' or 'false'.")
		}
		filter["is_fav"] = isFav
	}

	return filter, nil
}

func (h *BookmarkHandler) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	filter, err := h.buildBookmarkFilter(r, userID)
	if err != nil {
		utils.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var bookmarks []models.Bookmark
	collection := h.db.Client().Database("markly").Collection("bookmarks")

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Error().Err(err).Msg("Error finding bookmarks")
		http.Error(w, "Failed to retrieve bookmarks", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var bm models.Bookmark
		if err := cursor.Decode(&bm); err != nil {
			log.Error().Err(err).Msg("Decode error for bookmark")
			continue
		}
		bookmarks = append(bookmarks, bm)
	}

	if err := cursor.Err(); err != nil {
		log.Error().Err(err).Msg("Cursor error during bookmarks iteration")
		http.Error(w, "Error processing bookmarks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmarks)
}

func (h *BookmarkHandler) AddBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var reqBody models.AddBookmarkRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Error().Err(err).Msg("Error decoding request body for AddBookmark")
		utils.SendJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Debug().Interface("request_body", reqBody).Msg("Received bookmark request")

	if reqBody.URL == "" || reqBody.Title == "" {
		log.Error().Msg("URL and Title are required for AddBookmark")
		http.Error(w, "URL and Title are required", http.StatusBadRequest)
		return
	}

	var (
		tagsObjectIDs        []primitive.ObjectID
		collectionsObjectIDs []primitive.ObjectID
		categoryObjectIDPtr  *primitive.ObjectID
	)

	for _, tagIDStr := range reqBody.Tags {
		if tagIDStr == "" {
			continue
		}
		objID, err := primitive.ObjectIDFromHex(tagIDStr)
		if err != nil {
			log.Error().Err(err).Str("tag_id", tagIDStr).Msg("Invalid tag ID format")
			http.Error(w, "Invalid tag ID format: "+tagIDStr, http.StatusBadRequest)
			return
		}
		tagsObjectIDs = append(tagsObjectIDs, objID)
	}

	for _, colIDStr := range reqBody.Collections {
		if colIDStr == "" {
			continue
		}
		objID, err := primitive.ObjectIDFromHex(colIDStr)
		if err != nil {
			log.Error().Err(err).Str("collection_id", colIDStr).Msg("Invalid collection ID format")
			http.Error(w, "Invalid collection ID format: "+colIDStr, http.StatusBadRequest)
			return
		}
		collectionsObjectIDs = append(collectionsObjectIDs, objID)
	}

	if reqBody.CategoryID != nil && *reqBody.CategoryID != "" {
		catID, err := primitive.ObjectIDFromHex(*reqBody.CategoryID)
		if err != nil {
			log.Error().Err(err).Str("category_id", *reqBody.CategoryID).Msg("Invalid category ID format")
			http.Error(w, "Invalid category ID format: "+*reqBody.CategoryID, http.StatusBadRequest)
			return
		}
		categoryObjectIDPtr = &catID
	}

	if err := utils.ValidateReferences(h.db, userID, tagsObjectIDs, collectionsObjectIDs, categoryObjectIDPtr); err != nil {
		log.Error().Err(err).Msg("Reference validation failed for AddBookmark")
		http.Error(w, "Invalid reference: "+err.Error(), http.StatusBadRequest)
		return
	}

	bm := models.Bookmark{
		ID:            primitive.NewObjectID(),
		CreatedAt:     primitive.NewDateTimeFromTime(time.Now()),
		UserID:        userID,
		URL:           reqBody.URL,
		Title:         reqBody.Title,
		Summary:       reqBody.Summary,
		TagsID:        tagsObjectIDs,
		CollectionsID: collectionsObjectIDs,
		CategoryID:    categoryObjectIDPtr,
		IsFav:         reqBody.IsFav,
	}

	log.Debug().Interface("bookmark", bm).Msg("Creating bookmark")

	collection := h.db.Client().Database("markly").Collection("bookmarks")
	result, err := collection.InsertOne(context.Background(), bm)
	if err != nil {
		log.Error().Err(err).Msg("Error inserting bookmark")
		http.Error(w, "Failed to add bookmark", http.StatusInternalServerError)
		return
	}

	bm.ID = result.InsertedID.(primitive.ObjectID)

	log.Info().Str("bookmark_id", bm.ID.Hex()).Msg("Successfully created bookmark")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bm)
}


func (h *BookmarkHandler) GetBookmarkByID(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	bookmarkID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	filter := bson.M{"_id": bookmarkID, "user_id": userID}

	var bm models.Bookmark
	collection := h.db.Client().Database("markly").Collection("bookmarks")

	err = collection.FindOne(context.Background(), filter).Decode(&bm)
	if err == mongo.ErrNoDocuments {
		utils.SendJSONError(w, "Bookmark not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Error finding bookmark by ID")
		utils.SendJSONError(w, "Failed to retrieve bookmark", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, bm)
}

func (h *BookmarkHandler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	bookmarkID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	filter := bson.M{"_id": bookmarkID, "user_id": userID}

	collection := h.db.Client().Database("markly").Collection("bookmarks")
	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Error deleting bookmark")
		utils.SendJSONError(w, "Failed to delete bookmark", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		log.Warn().Str("bookmark_id", bookmarkID.Hex()).Msg("Bookmark not found or not authorized to delete")
		utils.SendJSONError(w, "Bookmark not found or not authorized to delete", http.StatusNotFound)
		return
	}

	log.Info().Str("bookmark_id", bookmarkID.Hex()).Msg("Bookmark deleted successfully")
	w.WriteHeader(http.StatusNoContent)
}

func (h *BookmarkHandler) buildUpdateFields(updatePayload models.UpdateBookmarkRequestBody, userID primitive.ObjectID) (bson.M, error) {
	updateFields := bson.M{}

	if updatePayload.URL != nil {
		updateFields["url"] = *updatePayload.URL
	}
	if updatePayload.Title != nil {
		updateFields["title"] = *updatePayload.Title
	}
	if updatePayload.Summary != nil {
		updateFields["summary"] = *updatePayload.Summary
	}

	// Handle Tags
	if updatePayload.Tags != nil {
		var tagsObjectIDs []primitive.ObjectID
		for _, tagIDStr := range *updatePayload.Tags {
			if tagIDStr == "" {
				continue
			}
			objID, err := primitive.ObjectIDFromHex(tagIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid tag ID format: %s", tagIDStr)
			}
			tagsObjectIDs = append(tagsObjectIDs, objID)
		}
		if err := utils.ValidateReferences(h.db, userID, tagsObjectIDs, nil, nil); err != nil {
			return nil, fmt.Errorf("invalid tag reference: %w", err)
		}
		updateFields["tagsid"] = tagsObjectIDs
	}

	// Handle Collections
	if updatePayload.Collections != nil {
		var collectionsObjectIDs []primitive.ObjectID
		for _, colIDStr := range *updatePayload.Collections {
			if colIDStr == "" {
				continue
			}
			objID, err := primitive.ObjectIDFromHex(colIDStr)
			if err != nil {
				return nil, fmt.Errorf("invalid collection ID format: %s", colIDStr)
			}
			collectionsObjectIDs = append(collectionsObjectIDs, objID)
		}
		if err := utils.ValidateReferences(h.db, userID, nil, collectionsObjectIDs, nil); err != nil {
			return nil, fmt.Errorf("invalid collection reference: %w", err)
		}
		updateFields["collectionsid"] = collectionsObjectIDs
	}

	// Handle CategoryID
	if updatePayload.CategoryID != nil {
		var categoryObjectIDPtr *primitive.ObjectID
		if *updatePayload.CategoryID == "" {
			// Frontend explicitly sent an empty string, meaning clear the category
			categoryObjectIDPtr = nil
		} else {
			// Attempt to convert the string to ObjectID
			objID, err := primitive.ObjectIDFromHex(*updatePayload.CategoryID)
			if err != nil {
				return nil, fmt.Errorf("invalid category ID format: %w", err)
			}
			categoryObjectIDPtr = &objID
		}

		if err := utils.ValidateReferences(h.db, userID, nil, nil, categoryObjectIDPtr); err != nil {
			return nil, fmt.Errorf("invalid category reference: %w", err)
		}
		updateFields["categoryid"] = categoryObjectIDPtr
	}

	if updatePayload.IsFav != nil {
		updateFields["is_fav"] = *updatePayload.IsFav
	}

	return updateFields, nil
}

func (h *BookmarkHandler) UpdateBookmark(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	bookmarkID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	var updatePayload models.UpdateBookmarkRequestBody
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		log.Error().Err(err).Msg("Invalid JSON for UpdateBookmark")
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	updateFields, err := h.buildUpdateFields(updatePayload, userID)
	if err != nil {
		log.Error().Err(err).Msg("Error building update fields for bookmark")
		utils.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(updateFields) == 0 {
		log.Warn().Msg("No valid fields provided for bookmark update")
		utils.SendJSONError(w, "No valid fields provided for update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := h.db.Client().Database("markly").Collection("bookmarks")
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Error updating bookmark")
		utils.SendJSONError(w, "Failed to update bookmark", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		log.Warn().Str("bookmark_id", bookmarkID.Hex()).Msg("Bookmark not found or not authorized to update")
		utils.SendJSONError(w, "Bookmark not found or not authorized to update", http.StatusNotFound)
		return
	}

	var updatedBookmark models.Bookmark
	err = collection.FindOne(context.Background(), filter).Decode(&updatedBookmark)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Msg("Error fetching updated bookmark")
		utils.SendJSONError(w, "Failed to retrieve updated bookmark", http.StatusInternalServerError)
		return
	}

	log.Info().Str("bookmark_id", bookmarkID.Hex()).Msg("Bookmark updated successfully")
	utils.RespondWithJSON(w, http.StatusOK, updatedBookmark)
}
