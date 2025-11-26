package repositories

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils"
)

type TrendingRepository interface {
	Create(ctx context.Context, item *models.TrendingItem) (*models.TrendingItem, error)
	FindByName(ctx context.Context, name string) (*models.TrendingItem, error)
	Update(ctx context.Context, name string, updateFields bson.M) (*mongo.UpdateResult, error)
	FindAll(ctx context.Context) ([]models.TrendingItem, error)
}

type trendingRepository struct {
	db database.Service
}

func NewTrendingRepository(db database.Service) TrendingRepository {
	return &trendingRepository{db: db}
}

func (r *trendingRepository) Create(ctx context.Context, item *models.TrendingItem) (*models.TrendingItem, error) {
	queryType := "create"
	repository := "trending"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("trending_items")
	_, err := collection.InsertOne(ctx, item)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		log.Error().Err(err).Str("name", item.Name).Msg("Failed to insert trending item into database")
		return nil, fmt.Errorf("failed to create trending item: %w", err)
	}
	return item, nil
}

func (r *trendingRepository) FindByName(ctx context.Context, name string) (*models.TrendingItem, error) {
	queryType := "findByName"
	repository := "trending"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("trending_items")
	var item models.TrendingItem
	err := collection.FindOne(ctx, bson.M{"name": name}).Decode(&item)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		return nil, err
	}
	return &item, nil
}

func (r *trendingRepository) Update(ctx context.Context, name string, updateFields bson.M) (*mongo.UpdateResult, error) {
	queryType := "update"
	repository := "trending"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("trending_items")
	update := bson.M{"$set": updateFields}
	result, err := collection.UpdateOne(ctx, bson.M{"name": name}, update)
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		log.Error().Err(err).Str("name", name).Msg("Error updating trending item")
		return nil, fmt.Errorf("failed to update trending item: %w", err)
	}
	return result, nil
}

func (r *trendingRepository) FindAll(ctx context.Context) ([]models.TrendingItem, error) {
	queryType := "findAll"
	repository := "trending"
	status := "success"
	timer := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		utils.DBQueryDurationSeconds.WithLabelValues(queryType, repository, status).Observe(v)
	}))
	defer timer.ObserveDuration()

	collection := r.db.Client().Database("markly").Collection("trending_items")
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		log.Error().Err(err).Msg("Failed to find all trending items")
		return nil, fmt.Errorf("failed to find all trending items: %w", err)
	}
	defer cursor.Close(ctx)

	var items []models.TrendingItem
	if err = cursor.All(ctx, &items); err != nil {
		status = "error"
		utils.DBQueryErrorsTotal.WithLabelValues(queryType, repository).Inc()
		log.Error().Err(err).Msg("Failed to decode trending items")
		return nil, fmt.Errorf("failed to decode trending items: %w", err)
	}
	return items, nil
}
