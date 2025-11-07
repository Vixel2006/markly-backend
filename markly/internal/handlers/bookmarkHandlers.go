package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
)

type BookmarkHandler struct {
	db database.Service
}

func NewBookmarksHandler(db database.Service) *BookmarkHandler {
	return &BookmarkHandler{db: db}
}

// parseObjectIDs helper function
func parseObjectIDs(idsStr string) ([]primitive.ObjectID, error) {
	var objectIDs []primitive.ObjectID
	if idsStr == "" {
		return objectIDs, nil
	}
	idStrings := strings.Split(idsStr, ",")
	for _, idStr := range idStrings {
		objID, err := primitive.ObjectIDFromHex(strings.TrimSpace(idStr))
		if err != nil {
			return nil, err
		}
		objectIDs = append(objectIDs, objID)
	}
	return objectIDs, nil
}

func (h *BookmarkHandler) GetBookmarks(w http.ResponseWriter, r *http.Request) {
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

	filter := bson.M{"user_id": userID}

	tagsParam := r.URL.Query().Get("tags")
	if tagsParam != "" {
		tagsIDs, err := parseObjectIDs(tagsParam)
		if err != nil {
			http.Error(w, "Invalid tags ID format. Tags must be comma-separated hexadecimal ObjectIDs.", http.StatusBadRequest)
			return
		}
		filter["tagsid"] = bson.M{"$in": tagsIDs}
	}

	categoryParam := r.URL.Query().Get("category")
	if categoryParam != "" {
		categoryID, err := primitive.ObjectIDFromHex(categoryParam)
		if err != nil {
			http.Error(w, "Invalid category ID format. Category must be a hexadecimal ObjectID.", http.StatusBadRequest)
			return
		}
		filter["categoryid"] = categoryID
	}

	collectionsParam := r.URL.Query().Get("collections")
	if collectionsParam != "" {
		collectionIDs, err := parseObjectIDs(collectionsParam)
		if err != nil {
			http.Error(w, "Invalid collections ID format. Collections must be comma-separated hexadecimal ObjectIDs.", http.StatusBadRequest)
			return
		}
		filter["collectionsid"] = bson.M{"$in": collectionIDs}
	}

	isFavParam := r.URL.Query().Get("isFav")
	if isFavParam != "" {
		isFav, err := strconv.ParseBool(isFavParam)
		if err != nil {
			http.Error(w, "Invalid isFav format. Must be 'true' or 'false'.", http.StatusBadRequest)
			return
		}
		filter["is_fav"] = isFav
	}

	var bookmarks []models.Bookmark
	collection := h.db.Client().Database("markly").Collection("bookmarks")

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error finding bookmarks: %v", err)
		http.Error(w, "Failed to retrieve bookmarks", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var bm models.Bookmark
		if err := cursor.Decode(&bm); err != nil {
			log.Printf("Decode error for bookmark: %v", err)
			continue
		}
		bookmarks = append(bookmarks, bm)
	}

	if err := cursor.Err(); err != nil {
		log.Printf("Cursor error during bookmarks iteration: %v", err)
		http.Error(w, "Error processing bookmarks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmarks)
}

func (h *BookmarkHandler) AddBookmark(w http.ResponseWriter, r *http.Request) {
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

	var reqBody models.AddBookmarkRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	log.Printf("Received bookmark request: %+v", reqBody)

	if reqBody.URL == "" || reqBody.Title == "" {
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
			log.Printf("Invalid tag ID format: %s, error: %v", tagIDStr, err)
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
			log.Printf("Invalid collection ID format: %s, error: %v", colIDStr, err)
			http.Error(w, "Invalid collection ID format: "+colIDStr, http.StatusBadRequest)
			return
		}
		collectionsObjectIDs = append(collectionsObjectIDs, objID)
	}

	if reqBody.CategoryID != nil && *reqBody.CategoryID != "" {
		catID, err := primitive.ObjectIDFromHex(*reqBody.CategoryID)
		if err != nil {
			log.Printf("Invalid category ID format: %s, error: %v", *reqBody.CategoryID, err)
			http.Error(w, "Invalid category ID format: "+*reqBody.CategoryID, http.StatusBadRequest)
			return
		}
		categoryObjectIDPtr = &catID
	}

	if err := h.validateReferences(userID, tagsObjectIDs, collectionsObjectIDs, categoryObjectIDPtr); err != nil {
		log.Printf("Reference validation failed: %v", err)
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

	log.Printf("Creating bookmark: %+v", bm)

	collection := h.db.Client().Database("markly").Collection("bookmarks")
	result, err := collection.InsertOne(context.Background(), bm)
	if err != nil {
		log.Printf("Error inserting bookmark: %v", err)
		http.Error(w, "Failed to add bookmark", http.StatusInternalServerError)
		return
	}

	bm.ID = result.InsertedID.(primitive.ObjectID)

	log.Printf("Successfully created bookmark with ID: %s", bm.ID.Hex())

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bm)
}

func (h *BookmarkHandler) validateReferences(userID primitive.ObjectID, tagIDs []primitive.ObjectID, collectionIDs []primitive.ObjectID, categoryID *primitive.ObjectID) error {
	ctx := context.Background()

	// Validate tags
	if len(tagIDs) > 0 {
		tagsCollection := h.db.Client().Database("markly").Collection("tags")
		count, err := tagsCollection.CountDocuments(ctx, bson.M{
			"_id":     bson.M{"$in": tagIDs},
			"user_id": userID,
		})
		if err != nil {
			return err
		}
		if count != int64(len(tagIDs)) {
			return errors.New("one or more tags not found or do not belong to user")
		}
	}

	if len(collectionIDs) > 0 {
		collectionsCollection := h.db.Client().Database("markly").Collection("collections")
		count, err := collectionsCollection.CountDocuments(ctx, bson.M{
			"_id":     bson.M{"$in": collectionIDs},
			"user_id": userID,
		})
		if err != nil {
			return err
		}
		if count != int64(len(collectionIDs)) {
			return errors.New("one or more collections not found or do not belong to user")
		}
	}

	// Validate category
	if categoryID != nil {
		categoriesCollection := h.db.Client().Database("markly").Collection("categories")
		count, err := categoriesCollection.CountDocuments(ctx, bson.M{
			"_id":     *categoryID,
			"user_id": userID,
		})
		if err != nil {
			return err
		}
		if count == 0 {
			return errors.New("category not found or does not belong to user")
		}
	}

	return nil
}

func (h *BookmarkHandler) GetBookmarkByID(w http.ResponseWriter, r *http.Request) {
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

	bookmarkID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID format", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": bookmarkID, "user_id": userID}

	var bm models.Bookmark
	collection := h.db.Client().Database("markly").Collection("bookmarks")

	err = collection.FindOne(context.Background(), filter).Decode(&bm)
	if err == mongo.ErrNoDocuments {
		http.Error(w, "Bookmark not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error finding bookmark by ID %s: %v", idStr, err)
		http.Error(w, "Failed to retrieve bookmark", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bm)
}

func (h *BookmarkHandler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
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

	bookmarkID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID format", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": bookmarkID, "user_id": userID}

	collection := h.db.Client().Database("markly").Collection("bookmarks")
	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Error deleting bookmark %s: %v", idStr, err)
		http.Error(w, "Failed to delete bookmark", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		http.Error(w, "Bookmark not found or not authorized to delete", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BookmarkHandler) UpdateBookmark(w http.ResponseWriter, r *http.Request) {
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

	bookmarkID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid bookmark ID format", http.StatusBadRequest)
		return
	}

	var updatePayload models.UpdateBookmarkRequestBody
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

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
				http.Error(w, "Invalid tag ID format: "+tagIDStr, http.StatusBadRequest)
				return
			}
			tagsObjectIDs = append(tagsObjectIDs, objID)
		}
		if err := h.validateReferences(userID, tagsObjectIDs, nil, nil); err != nil {
			http.Error(w, "Invalid tag reference: "+err.Error(), http.StatusBadRequest)
			return
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
				http.Error(w, "Invalid collection ID format: "+colIDStr, http.StatusBadRequest)
				return
			}
			collectionsObjectIDs = append(collectionsObjectIDs, objID)
		}
		if err := h.validateReferences(userID, nil, collectionsObjectIDs, nil); err != nil {
			http.Error(w, "Invalid collection reference: "+err.Error(), http.StatusBadRequest)
			return
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
				http.Error(w, "Invalid category ID format: "+err.Error(), http.StatusBadRequest)
				return
			}
			categoryObjectIDPtr = &objID
		}

		if err := h.validateReferences(userID, nil, nil, categoryObjectIDPtr); err != nil {
			http.Error(w, "Invalid category reference: "+err.Error(), http.StatusBadRequest)
			return
		}
		updateFields["categoryid"] = categoryObjectIDPtr
	}

	if updatePayload.IsFav != nil {
		updateFields["is_fav"] = *updatePayload.IsFav
	}

	if len(updateFields) == 0 {
		http.Error(w, "No valid fields provided for update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := h.db.Client().Database("markly").Collection("bookmarks")
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error updating bookmark %s: %v", idStr, err)
		http.Error(w, "Failed to update bookmark", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Bookmark not found or not authorized to update", http.StatusNotFound)
		return
	}

	var updatedBookmark models.Bookmark
	err = collection.FindOne(context.Background(), filter).Decode(&updatedBookmark)
	if err == mongo.ErrNoDocuments {
		http.Error(w, "Updated bookmark not found after update (internal error)", http.StatusInternalServerError)
		return
	}
	if err != nil {
		log.Printf("Error fetching updated bookmark %s: %v", idStr, err)
		http.Error(w, "Failed to retrieve updated bookmark", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedBookmark)
}
