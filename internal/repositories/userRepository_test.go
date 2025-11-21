package repositories

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"markly/internal/database"
	"markly/internal/models"
)

func TestUserRepository(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	db := database.New()
	defer db.Close()

	userRepo := NewUserRepository(db)

	t.Run("Create and Get User", func(t *testing.T) {
		user := &models.User{
			ID:       primitive.NewObjectID(),
			Username: "testuser",
			Email:    "test@example.com",
			Password: "password",
		}

		createdUser, err := userRepo.Create(context.Background(), user)
		assert.NoError(t, err)
		assert.NotNil(t, createdUser)

		foundUser, err := userRepo.FindByID(context.Background(), createdUser.ID)
		assert.NoError(t, err)
		assert.NotNil(t, foundUser)
		assert.Equal(t, createdUser.ID, foundUser.ID)

		_, err = userRepo.Delete(context.Background(), createdUser.ID)
		assert.NoError(t, err)
	})
}
