package handlers

import (
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

type BookmarkHandler struct {
	db database.Service
}

func NewBookmarksHandler(db database.Service) *BookmarkHandler {
	return &BookmarkHandler{db: db}
}

func (h *BookmarkHandler) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	var bookmarks []models.Bookmark
	collection := h.db.Client().Database("markly").Collection("bookmarks")
	cursor, err := collection.Find(context.Background(), bson.M{})

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var bm models.Bookmark

		if err := cursor.Decode(&bm); err != nil {
			log.Println("Decode error: ", err)
			continue
		}
		bookmarks = append(bookmarks, bm)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
  json.NewEncoder(w).Encode(bookmarks)
}

func (h *BookmarkHandler) AddBookmark(w http.ResponseWriter, r *http.Request) {
	var bm models.Bookmark

	if err := json.NewDecoder(r.Body).Decode(&bm); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	bm.ID = primitive.NewObjectID()
	bm.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

	collection := h.db.Client().Database("markly").Collection("bookmarks")
	_, err := collection.InsertOne(context.Background(), bm)

	if err != nil {
		http.Error(w, "Failed to insert bookmark", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(bm)
}

func (h *BookmarkHandler) GetBookmarkByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	collection := h.db.Client().Database("markly").Collection("bookmarks")

	id, err := primitive.ObjectIDFromHex(idStr)

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": id}

	var bm models.Bookmark

	err = collection.FindOne(context.Background(), filter).Decode(&bm)

	if err != nil {
		http.Error(w, "Cannot find the bookmark.", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bm)
}

func (h *BookmarkHandler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	collection := h.db.Client().Database("markly").Collection("bookmarks")

	id, err := primitive.ObjectIDFromHex(idStr)

	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": id}

	delete_result, err := collection.DeleteOne(context.Background(), filter)

	if err != nil {
		http.Error(w, "Cannot delete the bookmark", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(delete_result)
}

func (h *BookmarkHandler) UpdateBookmark(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Bookmark updated"})
}

func (h *BookmarkHandler) GetBookmarksByTags(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Bookmarks by tags"})
}

