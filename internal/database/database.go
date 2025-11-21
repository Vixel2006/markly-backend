package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog/log"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service interface {
	Health() map[string]string
	Client() *mongo.Client
}

type service struct {
	db *mongo.Client
}

var (
	host = os.Getenv("BLUEPRINT_DB_HOST")
	port = os.Getenv("BLUEPRINT_DB_PORT")
	//database = os.Getenv("BLUEPRINT_DB_DATABASE")
)

func New() Service {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(fmt.Sprintf("mongodb://%s:%s", host, port)))

	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	return &service{
		db: client,
	}
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.Ping(ctx, nil)
	if err != nil {
		log.Error().Err(err).Msg("Database health check failed")
		return map[string]string{
			"message": "db down",
			"error":   err.Error(),
		}
	}

	return map[string]string{
		"message": "It's healthy",
	}
}

func (s *service) Client() *mongo.Client {
	return s.db
}
