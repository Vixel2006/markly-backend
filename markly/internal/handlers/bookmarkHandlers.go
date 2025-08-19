package handlers

import (
	"context"
	"encoding/json"
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

type AddBookmarkRequestBody struct {
	URL         string   `json:"url"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Collections []string `json:"collections,omitempty"`
	Category    *string  `json:"category,omitempty"`
}

type BookmarkHandler struct {
	db database.Service
}

func NewBookmarksHandler(db database.Service) *BookmarkHandler {
	return &BookmarkHandler{db: db}
}

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
		// Query on TagsID field
		filter["tagsid"] = bson.M{"$in": tagsIDs}
	}

	categoryParam := r.URL.Query().Get("category")
	if categoryParam != "" {
		categoryID, err := primitive.ObjectIDFromHex(categoryParam)
		if err != nil {
			http.Error(w, "Invalid category ID format. Category must be a hexadecimal ObjectID.", http.StatusBadRequest)
			return
		}
		// Query on CategoryID field
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

	var reqBody AddBookmarkRequestBody
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

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
		objID, err := primitive.ObjectIDFromHex(tagIDStr)
		if err != nil {
			http.Error(w, "Invalid tag ID format: "+tagIDStr, http.StatusBadRequest)
			return
		}
		tagsObjectIDs = append(tagsObjectIDs, objID)
	}

	for _, colIDStr := range reqBody.Collections {
		objID, err := primitive.ObjectIDFromHex(colIDStr)
		if err != nil {
			http.Error(w, "Invalid collection ID format: "+colIDStr, http.StatusBadRequest)
			return
		}
		collectionsObjectIDs = append(collectionsObjectIDs, objID)
	}

	if reqBody.Category != nil && *reqBody.Category != "" {
		catID, err := primitive.ObjectIDFromHex(*reqBody.Category)
		if err != nil {
			http.Error(w, "Invalid category ID format: "+*reqBody.Category, http.StatusBadRequest)
			return
		}
		categoryObjectIDPtr = &catID
	} else {
		categoryObjectIDPtr = nil
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
		IsFav:         false,
	}

	collection := h.db.Client().Database("markly").Collection("bookmarks")
	_, err = collection.InsertOne(context.Background(), bm)
	if err != nil {
		log.Printf("Error inserting bookmark: %v", err)
		http.Error(w, "Failed to add bookmark", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bm)
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

	w.WriteHeader(http.StatusNoContent) // 204 No Content for successful deletion
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

	var nbm models.BookmarkUpdate
	if err := json.NewDecoder(r.Body).Decode(&nbm); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	updateFields := bson.M{}

	if nbm.URL != nil {
		updateFields["url"] = *nbm.URL
	}
	if nbm.Title != nil {
		updateFields["title"] = *nbm.Title
	}
	if nbm.Summary != nil {
		updateFields["summary"] = *nbm.Summary
	}
	if nbm.TagsID != nil {
		updateFields["tags"] = *nbm.TagsID
	}
	if nbm.CollectionsID != nil {
		updateFields["collections"] = *nbm.CollectionsID
	}
	if nbm.CategoryID != nil {
		if (*nbm.CategoryID).IsZero() {
			updateFields["category"] = primitive.Null{}
		} else {
			updateFields["category"] = *nbm.CategoryID
		}
	}
	if nbm.IsFav != nil {
		updateFields["is_fav"] = *nbm.IsFav
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
