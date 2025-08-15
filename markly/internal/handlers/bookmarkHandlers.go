package handlers

import (
	"encoding/json"
	"net/http"

	_ "github.com/joho/godotenv/autoload"

	"markly/internal/database"
)

type BookmarkHandler struct {
	db *database.Service
}

func NewBookmarksHandler(db *database.Service) *BookmarkHandler {
	return &BookmarkHandler{db: db}
}

func (h *BookmarkHandler) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Get all bookmarks"})
}

func (h *BookmarkHandler) AddBookmark(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Bookmark added"})
}

func (h *BookmarkHandler) GetBookmarkByID(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Single bookmark"})
}

func (h *BookmarkHandler) DeleteBookmark(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Bookmark deleted"})
}

func (h *BookmarkHandler) UpdateBookmark(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Bookmark updated"})
}

func (h *BookmarkHandler) GetBookmarksByTags(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(map[string]string{"message": "Bookmarks by tags"})
}

