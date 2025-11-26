package services

import (
	"context"
	"sort"
	"time"

	"markly/internal/models"
	"markly/internal/repositories"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AnalyticsService struct {
	UserRepository     *repositories.UserRepository
	BookmarkRepository *repositories.BookmarkRepository
	TagRepository      *repositories.TagRepository
	TrendingRepository *repositories.TrendingRepository
}

func NewAnalyticsService(
	userRepo *repositories.UserRepository,
	bookmarkRepo *repositories.BookmarkRepository,
	tagRepo *repositories.TagRepository,
	trendingRepo *repositories.TrendingRepository,
) *AnalyticsService {
	return &AnalyticsService{
		UserRepository:     userRepo,
		BookmarkRepository: bookmarkRepo,
		TagRepository:      tagRepo,
		TrendingRepository: trendingRepo,
	}
}

func (s *AnalyticsService) GetUserGrowth(ctx context.Context, startDate, endDate time.Time) (map[string]int, error) {
	count, err := (*s.UserRepository).CountUsersCreatedBetween(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return map[string]int{"new_users": int(count)}, nil
}

func (s *AnalyticsService) GetBookmarkActivity(ctx context.Context, startDate, endDate time.Time) (map[string]int, error) {
	count, err := (*s.BookmarkRepository).CountBookmarksCreatedBetween(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return map[string]int{"new_bookmarks": int(count)}, nil
}

func (s *AnalyticsService) GetBookmarkEngagement(ctx context.Context, userID primitive.ObjectID) (map[string]interface{}, error) {
	favoriteCount, err := (*s.BookmarkRepository).CountFavoriteBookmarks(ctx, userID)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"favorite_bookmarks": int(favoriteCount)}, nil
}

func (s *AnalyticsService) GetTagTrends(ctx context.Context) ([]models.Tag, error) {
	tags, err := (*s.TagRepository).FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Sort tags by WeeklyCount in descending order
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].WeeklyCount > tags[j].WeeklyCount
	})

	return tags, nil
}

func (s *AnalyticsService) GetTrendingItems(ctx context.Context) ([]models.TrendingItem, error) {
	items, err := (*s.TrendingRepository).FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Sort trending items by Count in descending order
	sort.Slice(items, func(i, j int) bool {
		return items[i].Count > items[j].Count
	})

	return items, nil
}
