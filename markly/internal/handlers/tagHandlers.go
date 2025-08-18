package handlers

import (
	"strings"
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

type TagHandler struct {
	db database.Service
}

func NewTagHandler(db database.Service) *TagHandler {
	return &TagHandler{db: db}
}

func (h *TagHandler) AddTag(w http.ResponseWriter, r *http.Request) {
	var tag models.Tag

	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		http.Error(w, "Cannot find tag.", http.StatusBadRequest)
	}

	tag.ID = primitive.NewObjectID()
	tag.WeeklyCount = 0
	tag.PrevCount = 0
	tag.CreatedAt = primitive.NewDateFromTime(time.Now())

	collection := h.db.Client().Database("markly").Collection("tags")

	indexModel := mongo.IndexModel{
		Keys: bson.M{"Name": 1},
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.TODO(), indexModel)

	if err != nil {
		log.Fatal(err)
	}

	_, err = collection.InsertOne(context.TODO(), tag)

	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Println("Already Exists")
		} else {
			log.Fatal(err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tag)
}

func (h *TagHandler) GetTagsByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ids := r.URL.Query()["id"]

	type result struct {
		Tag Tag
		Err Error
	}

	collection := h.db.Client().Database("markly").Collection("tags")

	results := make([]result, len(ids))

	var wg sync.WaitGroup
	wg.Add(len(ids))

	for i, id := range ids {
		i, id := i, id
		go func() {
			defer wg.Done()
			err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&tag)

			if err != nil {
				results[i] = result{Err: err}
				return
			}

			results[i] = result{Tag: tag, Err: nil}
		}()
	}

	wg.Wait()


	tags := []Tag{}

	for _, r := range results {
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

