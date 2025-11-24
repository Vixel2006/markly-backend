package services

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/repositories"
	"markly/internal/utils"
)

type BookmarkService interface {
	GetBookmarks(ctx context.Context, userID primitive.ObjectID, r *http.Request) ([]models.Bookmark, error)
	AddBookmark(ctx context.Context, userID primitive.ObjectID, reqBody models.AddBookmarkRequestBody) (*models.Bookmark, error)
	GetBookmarkByID(ctx context.Context, userID, bookmarkID primitive.ObjectID) (*models.Bookmark, error)
	DeleteBookmark(ctx context.Context, userID, bookmarkID primitive.ObjectID) (bool, error)
	UpdateBookmark(ctx context.Context, userID, bookmarkID primitive.ObjectID, updatePayload models.UpdateBookmarkRequestBody) (*models.Bookmark, error)
}

type bookmarkServiceImpl struct {
	bookmarkRepo repositories.BookmarkRepository
	db           database.Service
}

func NewBookmarkService(bookmarkRepo repositories.BookmarkRepository, db database.Service) BookmarkService {
	return &bookmarkServiceImpl{bookmarkRepo: bookmarkRepo, db: db}
}

func (s *bookmarkServiceImpl) buildBookmarkFilter(r *http.Request, userID primitive.ObjectID) (bson.M, error) {
	log.Debug().Str("userID", userID.Hex()).Msg("Building bookmark filter")
	filter := bson.M{"user_id": userID}

	tagsParam := r.URL.Query().Get("tags")
	if tagsParam != "" {
		tagsIDs, err := utils.ParseObjectIDs(tagsParam)
		if err != nil {
			log.Warn().Err(err).Str("tagsParam", tagsParam).Msg("Invalid tags ID format")
			return nil, fmt.Errorf("invalid tags ID format. Tags must be comma-separated hexadecimal ObjectIDs.")
		}
		filter["tagsid"] = bson.M{"$in": tagsIDs}
	}

	categoryParam := r.URL.Query().Get("category")
	if categoryParam != "" {
		categoryID, err := primitive.ObjectIDFromHex(categoryParam)
		if err != nil {
			log.Warn().Err(err).Str("categoryParam", categoryParam).Msg("Invalid category ID format")
			return nil, fmt.Errorf("invalid category ID format. Category must be a hexadecimal ObjectID.")
		}
		filter["categoryid"] = categoryID
	}

	collectionsParam := r.URL.Query().Get("collections")
	if collectionsParam != "" {
		collectionIDs, err := utils.ParseObjectIDs(collectionsParam)
		if err != nil {
			log.Warn().Err(err).Str("collectionsParam", collectionsParam).Msg("Invalid collections ID format")
			return nil, fmt.Errorf("invalid collections ID format. Collections must be comma-separated hexadecimal ObjectIDs.")
		}
		filter["collectionsid"] = bson.M{"$in": collectionIDs}
	}

	isFavParam := r.URL.Query().Get("isFav")
	if isFavParam != "" {
		isFav, err := strconv.ParseBool(isFavParam)
		if err != nil {
			log.Warn().Err(err).Str("isFavParam", isFavParam).Msg("Invalid isFav format")
			return nil, fmt.Errorf("invalid isFav format. Must be 'true' or 'false'.")
		}
		filter["is_fav"] = isFav
	}
	log.Debug().Str("userID", userID.Hex()).Interface("filter", filter).Msg("Bookmark filter built successfully")
	return filter, nil
}

func (s *bookmarkServiceImpl) GetBookmarks(ctx context.Context, userID primitive.ObjectID, r *http.Request) ([]models.Bookmark, error) {
	log.Debug().Str("userID", userID.Hex()).Msg("Attempting to retrieve bookmarks")
	filter, err := s.buildBookmarkFilter(r, userID)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Msg("Failed to build bookmark filter")
		return nil, err
	}

	var limit int64 = 5
	page, err := strconv.Atoi(r.URL.Query().Get("page"))

	if err != nil {
		log.Error().Err(err).Msg("Page Query should be an integer")
		return nil, err
	}

	bookmarks, err := s.bookmarkRepo.Find(ctx, filter, limit, int64(page))
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Interface("filter", filter).Msg("Error finding bookmarks")
		return nil, err
	}

	log.Debug().Str("userID", userID.Hex()).Int("count", len(bookmarks)).Msg("Successfully retrieved bookmarks")
	return bookmarks, nil
}

func (s *bookmarkServiceImpl) AddBookmark(ctx context.Context, userID primitive.ObjectID, reqBody models.AddBookmarkRequestBody) (*models.Bookmark, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("reqBody", reqBody).Msg("Attempting to add bookmark")
	if reqBody.URL == "" || reqBody.Title == "" {
		log.Warn().Str("userID", userID.Hex()).Msg("URL and Title are required for adding bookmark")
		return nil, fmt.Errorf("URL and Title are required")
	}

	var (
		tagsObjectIDs        []primitive.ObjectID
		collectionsObjectIDs []primitive.ObjectID
		categoryObjectIDPtr  *primitive.ObjectID
	)

	for _, tagIDStr := range reqBody.Tags {
		if tagIDStr == "" {
			continue
		}
		objID, err := primitive.ObjectIDFromHex(tagIDStr)
		if err != nil {
			log.Warn().Err(err).Str("userID", userID.Hex()).Str("tagIDStr", tagIDStr).Msg("Invalid tag ID format during AddBookmark")
			return nil, fmt.Errorf("invalid tag ID format: %s", tagIDStr)
		}
		tagsObjectIDs = append(tagsObjectIDs, objID)
	}

	for _, colIDStr := range reqBody.Collections {
		if colIDStr == "" {
			continue
		}
		objID, err := primitive.ObjectIDFromHex(colIDStr)
		if err != nil {
			log.Warn().Err(err).Str("userID", userID.Hex()).Str("colIDStr", colIDStr).Msg("Invalid collection ID format during AddBookmark")
			return nil, fmt.Errorf("invalid collection ID format: %s", colIDStr)
		}
		collectionsObjectIDs = append(collectionsObjectIDs, objID)
	}

	if reqBody.CategoryID != nil && *reqBody.CategoryID != "" {
		catID, err := primitive.ObjectIDFromHex(*reqBody.CategoryID)
		if err != nil {
			log.Warn().Err(err).Str("userID", userID.Hex()).Str("categoryIDStr", *reqBody.CategoryID).Msg("Invalid category ID format during AddBookmark")
			return nil, fmt.Errorf("invalid category ID format: %s", *reqBody.CategoryID)
		}
		categoryObjectIDPtr = &catID
	}

	if err := utils.ValidateReferences(s.db.Client(), userID, tagsObjectIDs, collectionsObjectIDs, categoryObjectIDPtr); err != nil {
		log.Warn().Err(err).Str("userID", userID.Hex()).Msg("Invalid reference during AddBookmark")
		return nil, fmt.Errorf("invalid reference: %w", err)
	}

	bm := models.Bookmark{
		ID:            primitive.NewObjectID(),
		CreatedAt:     primitive.NewDateTimeFromTime(time.Now()),
		UserID:        userID,
		URL:           reqBody.URL,
		Title:         reqBody.Title,
		Summary:       reqBody.Summary,
		TagsID:        tagsObjectIDs,
		CollectionsID: collectionsObjectIDs,
		CategoryID:    categoryObjectIDPtr,
		IsFav:         reqBody.IsFav,
	}

	createdBookmark, err := s.bookmarkRepo.Create(ctx, &bm)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Msg("Error inserting bookmark")
		return nil, err
	}

	log.Info().Str("userID", userID.Hex()).Str("bookmarkID", createdBookmark.ID.Hex()).Msg("Bookmark added successfully")
	return createdBookmark, nil
}

func (s *bookmarkServiceImpl) GetBookmarkByID(ctx context.Context, userID, bookmarkID primitive.ObjectID) (*models.Bookmark, error) {
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Attempting to retrieve bookmark by ID")
	filter := bson.M{"_id": bookmarkID, "user_id": userID}

	bm, err := s.bookmarkRepo.FindOne(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Warn().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Bookmark not found")
			return nil, fmt.Errorf("bookmark not found")
		}
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Str("userID", userID.Hex()).Msg("Error finding bookmark by ID")
		return nil, fmt.Errorf("failed to retrieve bookmark")
	}
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Successfully retrieved bookmark by ID")
	return bm, nil
}

func (s *bookmarkServiceImpl) DeleteBookmark(ctx context.Context, userID, bookmarkID primitive.ObjectID) (bool, error) {
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Attempting to delete bookmark")
	filter := bson.M{"_id": bookmarkID, "user_id": userID}

	deleteResult, err := s.bookmarkRepo.DeleteOne(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Str("userID", userID.Hex()).Msg("Error deleting bookmark")
		return false, err
	}

	if deleteResult.DeletedCount == 0 {
		log.Warn().Str("bookmark_id", bookmarkID.Hex()).Str("userID", userID.Hex()).Msg("Bookmark not found or not authorized to delete")
		return false, fmt.Errorf("bookmark not found or not authorized to delete")
	}
	log.Info().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Bookmark deleted successfully")
	return true, nil
}

func (s *bookmarkServiceImpl) buildUpdateFields(updatePayload models.UpdateBookmarkRequestBody, userID primitive.ObjectID) (bson.M, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("updatePayload", updatePayload).Msg("Building update fields for bookmark")
	updateFields := bson.M{}

	if updatePayload.URL != nil {
		updateFields["url"] = *updatePayload.URL
	}
	if updatePayload.Title != nil {
		updateFields["title"] = *updatePayload.Title
	}
	if updatePayload.Summary != nil {
		updateFields["summary"] = *updatePayload.Summary
	}

	// Handle Tags
	if updatePayload.Tags != nil {
		var tagsObjectIDs []primitive.ObjectID
		for _, tagIDStr := range *updatePayload.Tags {
			if tagIDStr == "" {
				continue
			}
			objID, err := primitive.ObjectIDFromHex(tagIDStr)
			if err != nil {
				log.Warn().Err(err).Str("userID", userID.Hex()).Str("tagIDStr", tagIDStr).Msg("Invalid tag ID format during buildUpdateFields")
				return nil, fmt.Errorf("invalid tag ID format: %s", tagIDStr)
			}
			tagsObjectIDs = append(tagsObjectIDs, objID)
		}
		if err := utils.ValidateReferences(s.db.Client(), userID, tagsObjectIDs, nil, nil); err != nil {
			log.Warn().Err(err).Str("userID", userID.Hex()).Msg("Invalid tag reference during buildUpdateFields")
			return nil, fmt.Errorf("invalid tag reference: %w", err)
		}
		updateFields["tagsid"] = tagsObjectIDs
	}

	// Handle Collections
	if updatePayload.Collections != nil {
		var collectionsObjectIDs []primitive.ObjectID
		for _, colIDStr := range *updatePayload.Collections {
			if colIDStr == "" {
				continue
			}
			objID, err := primitive.ObjectIDFromHex(colIDStr)
			if err != nil {
				log.Warn().Err(err).Str("userID", userID.Hex()).Str("colIDStr", colIDStr).Msg("Invalid collection ID format during buildUpdateFields")
				return nil, fmt.Errorf("invalid collection ID format: %s", colIDStr)
			}
			collectionsObjectIDs = append(collectionsObjectIDs, objID)
		}
		if err := utils.ValidateReferences(s.db.Client(), userID, nil, collectionsObjectIDs, nil); err != nil {
			log.Warn().Err(err).Str("userID", userID.Hex()).Msg("Invalid collection reference during buildUpdateFields")
			return nil, fmt.Errorf("invalid collection reference: %w", err)
		}
		updateFields["collectionsid"] = collectionsObjectIDs
	}

	// Handle CategoryID
	if updatePayload.CategoryID != nil {
		var categoryObjectIDPtr *primitive.ObjectID
		if *updatePayload.CategoryID == "" {
			// Frontend explicitly sent an empty string, meaning clear the category
			categoryObjectIDPtr = nil
		} else {
			// Attempt to convert the string to ObjectID
			objID, err := primitive.ObjectIDFromHex(*updatePayload.CategoryID)
			if err != nil {
				log.Warn().Err(err).Str("userID", userID.Hex()).Str("categoryIDStr", *updatePayload.CategoryID).Msg("Invalid category ID format during buildUpdateFields")
				return nil, fmt.Errorf("invalid category ID format: %w", err)
			}
			categoryObjectIDPtr = &objID
		}

		if err := utils.ValidateReferences(s.db.Client(), userID, nil, nil, categoryObjectIDPtr); err != nil {
			log.Warn().Err(err).Str("userID", userID.Hex()).Msg("Invalid category reference during buildUpdateFields")
			return nil, fmt.Errorf("invalid category reference: %w", err)
		}
		updateFields["categoryid"] = categoryObjectIDPtr
	}

	if updatePayload.IsFav != nil {
		updateFields["is_fav"] = *updatePayload.IsFav
	}
	log.Debug().Str("userID", userID.Hex()).Interface("updateFields", updateFields).Msg("Bookmark update fields built successfully")
	return updateFields, nil
}

func (s *bookmarkServiceImpl) UpdateBookmark(ctx context.Context, userID, bookmarkID primitive.ObjectID, updatePayload models.UpdateBookmarkRequestBody) (*models.Bookmark, error) {
	log.Debug().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Interface("updatePayload", updatePayload).Msg("Attempting to update bookmark")
	updateFields, err := s.buildUpdateFields(updatePayload, userID)
	if err != nil {
		log.Error().Err(err).Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Failed to build update fields for bookmark")
		return nil, err
	}

	if len(updateFields) == 0 {
		log.Warn().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("No valid fields provided for bookmark update")
		return nil, fmt.Errorf("no valid fields provided for update")
	}

	filter := bson.M{"_id": bookmarkID, "user_id": userID}
	update := bson.M{"$set": updateFields}

	result, err := s.bookmarkRepo.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Str("userID", userID.Hex()).Msg("Error updating bookmark")
		return nil, err
	}

	if result.MatchedCount == 0 {
		log.Warn().Str("bookmark_id", bookmarkID.Hex()).Str("userID", userID.Hex()).Msg("Bookmark not found or not authorized to update")
		return nil, fmt.Errorf("bookmark not found or not authorized to update")
	}

	updatedBookmark, err := s.bookmarkRepo.FindOne(ctx, filter)
	if err != nil {
		log.Error().Err(err).Str("bookmark_id", bookmarkID.Hex()).Str("userID", userID.Hex()).Msg("Error fetching updated bookmark")
		return nil, fmt.Errorf("failed to retrieve updated bookmark")
	}
	log.Info().Str("userID", userID.Hex()).Str("bookmarkID", bookmarkID.Hex()).Msg("Bookmark updated successfully")
	return updatedBookmark, nil
}
