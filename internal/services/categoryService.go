package services

import (
	"context"
	"fmt"
	"strings"


	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils"
)

// CategoryService defines the interface for category-related business logic.
type CategoryService interface {
	AddCategory(ctx context.Context, userID primitive.ObjectID, category models.Category) (*models.Category, error)
	GetCategories(ctx context.Context, userID primitive.ObjectID) ([]models.Category, error)
	GetCategoryByID(ctx context.Context, userID, categoryID primitive.ObjectID) (*models.Category, error)
	DeleteCategory(ctx context.Context, userID, categoryID primitive.ObjectID) (bool, error)
	UpdateCategory(ctx context.Context, userID, categoryID primitive.ObjectID, updatePayload models.CategoryUpdate) (*models.Category, error)
}

// categoryServiceImpl implements the CategoryService interface.
type categoryServiceImpl struct {
	db database.Service
}

// NewCategoryService creates a new CategoryService.
func NewCategoryService(db database.Service) CategoryService {
	return &categoryServiceImpl{db: db}
}

func (s *categoryServiceImpl) AddCategory(ctx context.Context, userID primitive.ObjectID, category models.Category) (*models.Category, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("categoryName", category.Name).Msg("Attempting to add category")
	category.ID = primitive.NewObjectID()
	category.UserID = userID


	collection := s.db.Client().Database("markly").Collection("categories")

	if err := utils.CreateUniqueIndex(collection, bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}}, "Category name"); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			log.Warn().Err(err).Str("userID", userID.Hex()).Interface("categoryName", category.Name).Msg("Category name already exists during index creation")
			return nil, fmt.Errorf("category name already exists")
		} else {
			log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to create index for category")
			return nil, fmt.Errorf("failed to set up category collection")
		}
	}

	_, err := collection.InsertOne(ctx, category)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("userID", userID.Hex()).Interface("categoryName", category.Name).Msg("Category name already exists for this user")
			return nil, fmt.Errorf("category name already exists for this user")
		} else {
			log.Error().Err(err).Str("category_name", category.Name).Str("user_id", userID.Hex()).Msg("Failed to insert category")
			return nil, fmt.Errorf("failed to insert category")
		}
	}
	log.Info().Str("userID", userID.Hex()).Str("categoryID", category.ID.Hex()).Interface("categoryName", category.Name).Msg("Category added successfully")
	return &category, nil
}

func (s *categoryServiceImpl) GetCategories(ctx context.Context, userID primitive.ObjectID) ([]models.Category, error) {
	log.Debug().Str("userID", userID.Hex()).Msg("Attempting to retrieve categories")
	var categories []models.Category

	collection := s.db.Client().Database("markly").Collection("categories")

	filter := bson.M{"user_id": userID}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error finding categories")
		return nil, fmt.Errorf("error fetching categories")
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &categories); err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error decoding categories")
		return nil, fmt.Errorf("error decoding categories")
	}
	log.Debug().Str("userID", userID.Hex()).Int("count", len(categories)).Msg("Successfully retrieved categories")
	return categories, nil
}

func (s *categoryServiceImpl) GetCategoryByID(ctx context.Context, userID, categoryID primitive.ObjectID) (*models.Category, error) {
	log.Debug().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Attempting to retrieve category by ID")
	filter := bson.M{"_id": categoryID, "user_id": userID}

	var category models.Category
	collection := s.db.Client().Database("markly").Collection("categories")

	err := collection.FindOne(ctx, filter).Decode(&category)
	if err == mongo.ErrNoDocuments {
		log.Warn().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Category not found")
		return nil, fmt.Errorf("category not found")
	}
	if err != nil {
		log.Error().Err(err).Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Error finding category by ID")
		return nil, fmt.Errorf("failed to retrieve category")
	}
	log.Debug().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Successfully retrieved category by ID")
	return &category, nil
}

func (s *categoryServiceImpl) DeleteCategory(ctx context.Context, userID, categoryID primitive.ObjectID) (bool, error) {
	log.Debug().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Attempting to delete category")
	collection := s.db.Client().Database("markly").Collection("categories")

	filter := bson.M{"user_id": userID, "_id": categoryID}

	deleteResult, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to delete category")
		return false, fmt.Errorf("failed to delete category")
	}

	if deleteResult.DeletedCount == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Category not found or unauthorized to delete")
		return false, fmt.Errorf("category not found or unauthorized to delete")
	}
	log.Info().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Category deleted successfully")
	return true, nil
}

func (s *categoryServiceImpl) buildCategoryUpdateFields(updatePayload models.CategoryUpdate) (bson.M, error) {
	log.Debug().Interface("updatePayload", updatePayload).Msg("Building category update fields")
	updateFields := bson.M{}
	if updatePayload.Name != nil {
		updateFields["name"] = *updatePayload.Name
	}
	if updatePayload.Emoji != nil {
		updateFields["emoji"] = *updatePayload.Emoji
	}
	log.Debug().Interface("updateFields", updateFields).Msg("Category update fields built successfully")
	return updateFields, nil
}

func (s *categoryServiceImpl) UpdateCategory(ctx context.Context, userID, categoryID primitive.ObjectID, updatePayload models.CategoryUpdate) (*models.Category, error) {
	log.Debug().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Interface("updatePayload", updatePayload).Msg("Attempting to update category")
	updateFields, err := s.buildCategoryUpdateFields(updatePayload)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Failed to build category update fields")
		return nil, err
	}

	if len(updateFields) == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("No fields to update for category")
		return nil, fmt.Errorf("no fields to update")
	}

	filter := bson.M{"_id": categoryID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	collection := s.db.Client().Database("markly").Collection("categories")

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Err(err).Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Category name already exists for this user during update")
			return nil, fmt.Errorf("category name already exists for this user")
		}
		log.Error().Err(err).Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to update category")
		return nil, fmt.Errorf("failed to update category")
	}

	if result.MatchedCount == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Category not found or unauthorized to update")
		return nil, fmt.Errorf("category not found or unauthorized to update")
	}

	var updatedCategory models.Category
	err = collection.FindOne(ctx, filter).Decode(&updatedCategory)
	if err != nil {
		log.Error().Err(err).Str("category_id", categoryID.Hex()).Str("user_id", userID.Hex()).Msg("Failed to find updated category")
		return nil, fmt.Errorf("failed to retrieve the updated category")
	}
	log.Info().Str("userID", userID.Hex()).Str("categoryID", categoryID.Hex()).Msg("Category updated successfully")
	return &updatedCategory, nil
}
