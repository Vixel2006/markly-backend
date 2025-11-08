package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"markly/internal/models"
	"os"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

var apiKey = os.Getenv("API_KEY")

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
		title,
		url,
	)

	summary, err := llms.GenerateFromSinglePrompt(context.Background(), llm, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate summary from LLM: %w", err)
	}

	return summary, nil
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
			bm.Title,
			bm.URL,
			bm.Summary,
			bm.Category,
			bm.Tags)
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
]`, recentBookmarksStr)

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
