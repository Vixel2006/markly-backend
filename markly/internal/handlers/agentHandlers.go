package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"markly/internal/database"
	"markly/internal/models"
	"net/http"
	"os"

	"errors"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

var apiKey = os.Getenv("API_KEY")

type AgentHandler struct {
	db database.Service
}

func NewAgentHandler(db database.Service) *AgentHandler {
	return &AgentHandler{db: db}
}

func LLMSummarize(url, title string) (string, error) {
	if apiKey == "" {
		return "", errors.New("missing api key.")
	}

	llm, err := googleai.New(context.Background(), googleai.WithAPIKey(apiKey), googleai.WithDefaultModel("gemini-2.5-flash"))
	if err != nil {
		return "", fmt.Errorf("failed to create Google AI LLM: %w", err)
	}

	prompt := fmt.Sprintf(
		"You are a bookmark summarizer. Generate a concise summary in Markdown format. "+
			"Use headings, bullets when helpful. Return only Markdown.\n\nTitle: %s\nURL: %s",
		title, url,
	)

	summary, err := llms.GenerateFromSinglePrompt(context.Background(), llm, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary from LLM: %w", err)
	}

	return summary, nil
}

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

func (a *AgentHandler) GenerateSummary(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		sendJSONError(w, "Invalid user ID", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		sendJSONError(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	bookmarkID, err := primitive.ObjectIDFromHex(vars["id"])
	if err != nil {
		sendJSONError(w, "Invalid bookmark ID format", http.StatusBadRequest)
		return
	}

	// 1. Fetch bookmark
	var bookmark models.Bookmark
	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	err = a.db.Client().Database("markly").
		Collection("bookmarks").
		FindOne(context.Background(), filter).
		Decode(&bookmark)
	if err != nil {
		sendJSONError(w, "Bookmark not found", http.StatusNotFound)
		return
	}

	// 2. Call your LLM summarizer (langchain service or direct API)
	summary, err := LLMSummarize(bookmark.URL, bookmark.Title)
	if err != nil {
		log.Printf("Error generating summary for bookmark %s: %v", bookmarkID.Hex(), err)
		sendJSONError(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	// 3. Update DB
	update := bson.M{"$set": bson.M{"summary": summary}}
	_, err = a.db.Client().Database("markly").
		Collection("bookmarks").
		UpdateOne(context.Background(), filter, update)
	if err != nil {
		sendJSONError(w, "Failed to save summary", http.StatusInternalServerError)
		return
	}

	// 4. Return updated bookmark
	bookmark.Summary = summary
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bookmark)
}
