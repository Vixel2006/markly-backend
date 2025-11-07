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
	"strings"
	"time"

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

func LLMGenerateSuggestions(recentBookmarks []models.PromptBookmarkInfo) ([]models.AISuggestion, error) {
	if apiKey == "" {
		return nil, errors.New("missing api key")
	}

	llm, err := googleai.New(context.Background(), googleai.WithAPIKey(apiKey), googleai.WithDefaultModel("gemini-2.5-flash"))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google AI LLM: %w", err)
	}

	var recentBookmarksStr string
	for _, bm := range recentBookmarks {
		recentBookmarksStr += fmt.Sprintf("- Title: %s, URL: %s, Summary: %s, Category: %s, Tags: %v\n",
			bm.Title, bm.URL, bm.Summary, bm.Category, bm.Tags)
	}

	prompt := fmt.Sprintf(`You are an AI assistant that suggests new bookmarks based on a user's recent activity.
The user's recent bookmarks are:
%s

Based on these, suggest 3 new, distinct bookmarks that the user might find interesting.
For each suggestion, provide a URL, Title, a concise Summary (in Markdown format), a Category, a Collection, and a list of Tags.
The Category, Collection, and Tags should be single words or short phrases, similar to the user's existing ones.
Return ONLY the JSON array of objects, with no additional text or markdown formatting outside the JSON.
The JSON array should contain exactly 3 objects, each with the following structure:
{
  "url": "string",
  "title": "string",
  "summary": "string (Markdown)",
  "category": "string",
  "collection": "string",
  "tags": ["string", "string"]
}

Example of expected JSON output:
[
  {
    "url": "https://example.com/article1",
    "title": "Example Article One",
    "summary": "A summary of example article one.",
    "category": "Technology",
    "collection": "Reading List",
    "tags": ["AI", "Future"]
  },
  {
    "url": "https://example.com/article2",
    "title": "Example Article Two",
    "summary": "A summary of example article two.",
    "category": "Science",
    "collection": "Research",
    "tags": ["Physics", "Quantum"]
  },
  {
    "url": "https://example.com/article3",
    "title": "Example Article Three",
    "summary": "A summary of example article three.",
    "category": "History",
    "collection": "Learning",
    "tags": ["Ancient", "Civilizations"]
  }
]
`, recentBookmarksStr)

	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		llmResponse, err := llms.GenerateFromSinglePrompt(context.Background(), llm, prompt)
		if err != nil {
			return nil, fmt.Errorf("failed to generate suggestions from LLM on retry %d: %w", i+1, err)
		}

		if llmResponse == "" {
			log.Printf("LLM returned an empty response on retry %d", i+1)
			continue // Retry if empty
		}

		// Robustly remove markdown code block fences if present
		cleanedResponse := strings.TrimSpace(llmResponse)
		if strings.HasPrefix(cleanedResponse, "```json") {
			cleanedResponse = strings.TrimPrefix(cleanedResponse, "```json")
		}
		if strings.HasSuffix(cleanedResponse, "```") {
			cleanedResponse = strings.TrimSuffix(cleanedResponse, "```")
		}
		cleanedResponse = strings.TrimSpace(cleanedResponse)

		var suggestions []models.AISuggestion
		err = json.Unmarshal([]byte(cleanedResponse), &suggestions)
		if err != nil {
			log.Printf("LLM raw response (retry %d): %s", i+1, llmResponse)         // Log the raw response for debugging
			log.Printf("LLM cleaned response (retry %d): %s", i+1, cleanedResponse) // Log the cleaned response for debugging
			return nil, fmt.Errorf("failed to parse LLM response as JSON on retry %d: %w", i+1, err)
		}

		if len(suggestions) == 3 {
			return suggestions, nil // Success
		}
		log.Printf("LLM returned %d suggestions on retry %d, expected 3. Retrying...", len(suggestions), i+1)
	}

	return nil, errors.New("LLM failed to generate exactly 3 suggestions after multiple retries")
}

func (a *AgentHandler) GenerateAISuggestions(w http.ResponseWriter, r *http.Request) {
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

	// Fetch user's bookmarks from the last 7 days
	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	filter := bson.M{
		"user_id":    userID,
		"created_at": bson.M{"$gte": sevenDaysAgo},
	}

	cursor, err := a.db.Client().Database("markly").Collection("bookmarks").Find(context.Background(), filter)
	if err != nil {
		sendJSONError(w, "Failed to fetch recent bookmarks", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var recentBookmarks []models.Bookmark
	if err = cursor.All(context.Background(), &recentBookmarks); err != nil {
		sendJSONError(w, "Failed to decode recent bookmarks", http.StatusInternalServerError)
		return
	}

	// Fetch all categories for the user
	categoryCursor, err := a.db.Client().Database("markly").Collection("categories").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		sendJSONError(w, "Failed to fetch categories", http.StatusInternalServerError)
		return
	}
	defer categoryCursor.Close(context.Background())
	var categories []models.Category
	if err = categoryCursor.All(context.Background(), &categories); err != nil {
		sendJSONError(w, "Failed to decode categories", http.StatusInternalServerError)
		return
	}
	categoryMap := make(map[primitive.ObjectID]string)
	for _, cat := range categories {
		categoryMap[cat.ID] = cat.Name
	}

	// Fetch all collections for the user
	collectionCursor, err := a.db.Client().Database("markly").Collection("collections").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		sendJSONError(w, "Failed to fetch collections", http.StatusInternalServerError)
		return
	}
	defer collectionCursor.Close(context.Background())
	var collections []models.Collection
	if err = collectionCursor.All(context.Background(), &collections); err != nil {
		sendJSONError(w, "Failed to decode collections", http.StatusInternalServerError)
		return
	}
	collectionMap := make(map[primitive.ObjectID]string)
	for _, col := range collections {
		collectionMap[col.ID] = col.Name
	}

	// Fetch all tags for the user
	tagCursor, err := a.db.Client().Database("markly").Collection("tags").Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		sendJSONError(w, "Failed to fetch tags", http.StatusInternalServerError)
		return
	}
	defer tagCursor.Close(context.Background())
	var tags []models.Tag
	if err = tagCursor.All(context.Background(), &tags); err != nil {
		sendJSONError(w, "Failed to decode tags", http.StatusInternalServerError)
		return
	}
	tagMap := make(map[primitive.ObjectID]string)
	for _, tag := range tags {
		tagMap[tag.ID] = tag.Name
	}

	var promptBookmarks []models.PromptBookmarkInfo
	for _, bm := range recentBookmarks {
		var categoryName string
		if bm.CategoryID != nil {
			categoryName = categoryMap[*bm.CategoryID]
		}

		var collectionName string
		if len(bm.CollectionsID) > 0 {
			// For simplicity, taking the first collection name if multiple exist
			collectionName = collectionMap[bm.CollectionsID[0]]
		}

		var tagNames []string
		for _, tagID := range bm.TagsID {
			if tagName, ok := tagMap[tagID]; ok {
				tagNames = append(tagNames, tagName)
			}
		}

		promptBookmarks = append(promptBookmarks, models.PromptBookmarkInfo{
			URL:        bm.URL,
			Title:      bm.Title,
			Summary:    bm.Summary,
			Category:   categoryName,
			Collection: collectionName,
			Tags:       tagNames,
		})
	}

	if len(promptBookmarks) == 0 {
		sendJSONError(w, "No recent bookmarks found to generate suggestions from. Please add some bookmarks first.", http.StatusOK)
		return
	}

	// Generate suggestions using LLM
	suggestions, err := LLMGenerateSuggestions(promptBookmarks)
	if err != nil {
		log.Printf("Error generating AI suggestions: %v", err)
		sendJSONError(w, fmt.Sprintf("Failed to generate AI suggestions: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(suggestions)
}

type SummarizeURLRequest struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

func (a *AgentHandler) SummarizeURL(w http.ResponseWriter, r *http.Request) {
	var req SummarizeURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		sendJSONError(w, "URL is required", http.StatusBadRequest)
		return
	}

	summary, err := LLMSummarize(req.URL, req.Title)
	if err != nil {
		log.Printf("Error generating summary for URL %s: %v", req.URL, err)
		sendJSONError(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"summary": summary})
}
