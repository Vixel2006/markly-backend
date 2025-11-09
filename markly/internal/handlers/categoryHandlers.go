package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils"
)

type CategoryHandler struct {
	db database.Service
}

func NewCategoryHandler(db database.Service) *CategoryHandler {
	return &CategoryHandler{db: db}
}

func (h *CategoryHandler) AddCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		utils.SendJSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	category.ID = primitive.NewObjectID()
	category.UserID = userID

	collection := h.db.Client().Database("markly").Collection("categories")

	if err := utils.CreateUniqueIndex(collection, bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}}, "Category name"); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.SendJSONError(w, err.Error(), http.StatusConflict)
		} else {
			log.Printf("Failed to create index for category: %v", err)
			utils.SendJSONError(w, "Failed to set up category collection", http.StatusInternalServerError)
		}
		return
	}

	_, err = collection.InsertOne(context.Background(), category)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Println("Category name already exists for this user.")
			utils.SendJSONError(w, "Category name already exists for this user.", http.StatusConflict)
		} else {
			log.Printf("Failed to insert category: %v", err)
			utils.SendJSONError(w, "Failed to insert category", http.StatusInternalServerError)
		}
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, category)
}

func (h *CategoryHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	var categories []models.Category

	collection := h.db.Client().Database("markly").Collection("categories")

	filter := bson.M{"user_id": userID}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		log.Printf("Error finding categories: %v", err)
		utils.SendJSONError(w, "Error fetching categories", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &categories); err != nil {
		log.Printf("Error decoding categories: %v", err)
		utils.SendJSONError(w, "Error decoding categories", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, categories)
}

func (h *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	categoryID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	filter := bson.M{"_id": categoryID, "user_id": userID}

	var category models.Category
	collection := h.db.Client().Database("markly").Collection("categories")

	err = collection.FindOne(context.Background(), filter).Decode(&category)
	if err == mongo.ErrNoDocuments {
		utils.SendJSONError(w, "Category not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("Error finding category by ID %s: %v", categoryID.Hex(), err)
		utils.SendJSONError(w, "Failed to retrieve category", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, category)
}

func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	categoryID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	collection := h.db.Client().Database("markly").Collection("categories")

	filter := bson.M{"user_id": userID, "_id": categoryID}

	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Failed to delete category with ID %s for user %s: %v", categoryID.Hex(), userID.Hex(), err)
		utils.SendJSONError(w, "Failed to delete category", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		utils.SendJSONError(w, "Category not found or unauthorized", http.StatusNotFound)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, bson.M{"message": "Category deleted successfully", "deleted_count": deleteResult.DeletedCount})
}

func (h *CategoryHandler) buildCategoryUpdateFields(updatePayload models.CategoryUpdate) (bson.M, error) {
	updateFields := bson.M{}
	if updatePayload.Name != nil {
		updateFields["name"] = *updatePayload.Name
	}
	if updatePayload.Emoji != nil {
		updateFields["emoji"] = *updatePayload.Emoji
	}
	return updateFields, nil
}

func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.GetUserIDFromContext(w, r)
	if err != nil {
		return
	}

	categoryID, err := utils.GetObjectIDFromVars(w, r, "id")
	if err != nil {
		return
	}

	var updatePayload models.CategoryUpdate
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		utils.SendJSONError(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updateFields, err := h.buildCategoryUpdateFields(updatePayload)
	if err != nil {
		utils.SendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(updateFields) == 0 {
		utils.SendJSONError(w, "No fields to update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": categoryID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := h.db.Client().Database("markly").Collection("categories")

	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			utils.SendJSONError(w, "Category name already exists for this user.", http.StatusConflict)
			return
		}
		log.Printf("Failed to update category with ID %s for user %s: %v", categoryID.Hex(), userID.Hex(), err)
		utils.SendJSONError(w, "Failed to update category", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendJSONError(w, "Category not found or unauthorized to update", http.StatusNotFound)
		return
	}

	var updatedCategory models.Category
	err = collection.FindOne(context.Background(), filter).Decode(&updatedCategory)
	if err != nil {
		log.Printf("Failed to find updated category with ID %s for user %s: %v", categoryID.Hex(), userID.Hex(), err)
		utils.SendJSONError(w, "Failed to retrieve the updated category", http.StatusInternalServerError)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, updatedCategory)
}
