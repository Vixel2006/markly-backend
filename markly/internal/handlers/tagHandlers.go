package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"markly/internal/database"
	"markly/internal/models"
)

type TagHandler struct {
	db database.Service
}

func NewTagHandler(db database.Service) *TagHandler {
	return &TagHandler{db: db}
}

func (h *TagHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	var tag models.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		http.Error(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
		return
	}

	tag.ID = primitive.NewObjectID()
	tag.WeeklyCount = 0
	tag.PrevCount = 0
	tag.CreatedAt = primitive.NewDateTimeFromTime(time.Now())

	collection := h.db.Client().Database("markly").Collection("tags")

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"name": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Printf("Duplicate tag name: %v", err)
			http.Error(w, "Tag name already exists.", http.StatusConflict)
		} else {
			log.Printf("Failed to create index for tags: %v", err)
			http.Error(w, "Failed to set up tag collection", http.StatusInternalServerError)
		}
		return
	}

	_, err = collection.InsertOne(context.Background(), tag)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Println("Tag name already exists.")
			http.Error(w, "Tag name already exists.", http.StatusConflict)
		} else {
			log.Printf("Failed to insert tag: %v", err)
			http.Error(w, "Failed to insert tag", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tag)
}

func (h *TagHandler) GetTagsByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ids := r.URL.Query()["id"]

	if len(ids) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.Tag{})
		return
	}

	type result struct {
		Tag models.Tag
		Err error
	}

	collection := h.db.Client().Database("markly").Collection("tags")

	resultsChan := make(chan result, len(ids))
	var wg sync.WaitGroup
	wg.Add(len(ids))

	for _, idStr := range ids {
		idStr := idStr
		go func() {
			defer wg.Done()

			objID, err := primitive.ObjectIDFromHex(strings.TrimSpace(idStr))
			if err != nil {
				log.Printf("Invalid tag ID format: %s - %v", idStr, err)
				resultsChan <- result{Err: err}
				return
			}

			var tag models.Tag
			filter := bson.M{"_id": objID}
			err = collection.FindOne(ctx, filter).Decode(&tag)

			if err != nil {
				log.Printf("Error finding tag %s: %v", idStr, err)
				resultsChan <- result{Err: err}
				return
			}

			resultsChan <- result{Tag: tag, Err: nil}
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

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tags); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *TagHandler) DeleteTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	tagID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid tag ID format", http.StatusBadRequest)
		return
	}

	collection := h.db.Client().Database("markly").Collection("tags")
	filter := bson.M{"_id": tagID}

	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Failed to delete tag with ID %s: %v", idStr, err)
		http.Error(w, "Failed to delete tag", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bson.M{"message": "Tag deleted successfully", "deleted_count": deleteResult.DeletedCount})
}

func (h *TagHandler) UpdateTag(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	idStr := vars["id"]

	tagID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid tag ID format", http.StatusBadRequest)
		return
	}

	var updatePayload struct {
		Name        *string `json:"name,omitempty" bson:"name,omitempty"`
		WeeklyCount *int    `json:"weeklyCount,omitempty" bson:"weekly_count,omitempty"`
		PrevCount   *int    `json:"prevCount,omitempty" bson:"prev_count,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

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

	if len(updateFields) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": tagID}
	update := bson.M{"$set": updateFields}

	collection := h.db.Client().Database("markly").Collection("tags")

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			http.Error(w, "Tag name already exists.", http.StatusConflict)
		} else {
			log.Printf("Failed to update tag with ID %s: %v", idStr, err)
			http.Error(w, "Failed to update tag", http.StatusInternalServerError)
		}
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "Tag not found", http.StatusNotFound)
		return
	}

	var updatedTag models.Tag
	err = collection.FindOne(context.Background(), filter).Decode(&updatedTag)
	if err != nil {
		log.Printf("Failed to find updated tag with ID %s: %v", idStr, err)
		http.Error(w, "Failed to retrieve the updated tag", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedTag)
}
