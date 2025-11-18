package utils

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"markly/internal/database"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// parseObjectIDs helper function to parse comma-separated ObjectID strings
func ParseObjectIDs(idsStr string) ([]primitive.ObjectID, error) {
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

// ValidateReferences checks if the provided ObjectIDs for tags, collections, and categories exist and belong to the user.
func ValidateReferences(db database.Service, userID primitive.ObjectID, tagIDs []primitive.ObjectID, collectionIDs []primitive.ObjectID, categoryID *primitive.ObjectID) error {
	ctx := context.Background()

	// Validate tags
	if len(tagIDs) > 0 {
		tagsCollection := db.Client().Database("markly").Collection("tags")
		count, err := tagsCollection.CountDocuments(ctx, bson.M{
			"_id":     bson.M{"$in": tagIDs},
			"user_id": userID,
		})
		if err != nil {
			return err
		}
		if count != int64(len(tagIDs)) {
			return errors.New("one or more tags not found or do not belong to user")
		}
	}

	// Validate collections
	if len(collectionIDs) > 0 {
		collectionsCollection := db.Client().Database("markly").Collection("collections")
		count, err := collectionsCollection.CountDocuments(ctx, bson.M{
			"_id":     bson.M{"$in": collectionIDs},
			"user_id": userID,
		})
		if err != nil {
			return err
		}
		if count != int64(len(collectionIDs)) {
			return errors.New("one or more collections not found or do not belong to user")
		}
	}

	// Validate category
	if categoryID != nil {
		categoriesCollection := db.Client().Database("markly").Collection("categories")
		count, err := categoriesCollection.CountDocuments(ctx, bson.M{
			"_id":     *categoryID,
			"user_id": userID,
		})
		if err != nil {
			return err
		}
		if count == 0 {
			return errors.New("category not found or does not belong to user")
		}
	}

	return nil
}

// CreateUniqueIndex creates a unique index on the specified collection and keys.
// It returns an error if the index creation fails, including a specific error for duplicate keys.
func CreateUniqueIndex(collection *mongo.Collection, keys interface{}, fieldName string) error {
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(true),
	}

	_, err := collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return fmt.Errorf("%s already exists", fieldName)
		}
		return fmt.Errorf("failed to create index for %s: %w", fieldName, err)
	}
	return nil
}