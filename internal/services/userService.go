package services

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"markly/internal/models"
	"markly/internal/repositories"
	"markly/internal/utils"
)

// UserService defines the interface for user-related business logic.
type UserService interface {
	RegisterUser(ctx context.Context, user *models.User) (*models.User, error)
	LoginUser(ctx context.Context, creds *models.Login) (string, error)
	GetUserProfile(ctx context.Context, userID primitive.ObjectID) (*models.User, error)
	UpdateUserProfile(ctx context.Context, userID primitive.ObjectID, updatePayload *models.UserProfileUpdate) (*models.User, error)
	DeleteUser(ctx context.Context, userID primitive.ObjectID) error
	GetTotalUsers(ctx context.Context) (int64, error)
}

// userService implements UserService using a UserRepository.
type userService struct {
	userRepo        repositories.UserRepository
	totalUsersGauge prometheus.Gauge
}

// NewUserService creates a new UserService.
func NewUserService(userRepo repositories.UserRepository) UserService {
	s := &userService{
		userRepo: userRepo,
		totalUsersGauge: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "app_total_users",
			Help: "Total number of registered users in the application.",
		}),
	}
	go s.updateTotalUsersPeriodically()
	return s
}

func (s *userService) GetTotalUsers(ctx context.Context) (int64, error) {
	return s.userRepo.CountAll(ctx)
}

func (s *userService) updateTotalUsersPeriodically() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		count, err := s.GetTotalUsers(ctx)
		if err != nil {
			log.Error().Err(err).Msg("Error updating total users gauge")
		} else {
			s.totalUsersGauge.Set(float64(count))
		}
		cancel()
	}
}

func (s *userService) RegisterUser(ctx context.Context, user *models.User) (*models.User, error) {
	log.Debug().Str("email", user.Email).Msg("Attempting to register user")
	if user.Username == "" || user.Email == "" || user.Password == "" {
		log.Warn().Msg("Username, email, and password are required for registration")
		return nil, fmt.Errorf("username, email, and password are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		log.Error().Err(err).Msg("Failed to hash password during registration")
		return nil, fmt.Errorf("failed to hash password")
	}

	user.Password = string(hashedPassword)
	user.ID = primitive.NewObjectID()

	// ToDo: This logic should be in the repository
	// if err := utils.CreateUniqueIndex(collection, bson.M{"email": 1}, "Email"); err != nil {
	// 	if strings.Contains(err.Error(), "already exists") {
	// 		log.Warn().Err(err).Str("email", user.Email).Msg("Email already exists during index creation")
	// 		return nil, fmt.Errorf("email already exists")
	// 	}
	// 	log.Error().Err(err).Msg("Error creating unique email index")
	// 	return nil, fmt.Errorf("failed to set up database index")
	// }

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			log.Warn().Str("email", user.Email).Msg("Email already exists during user insertion")
			return nil, fmt.Errorf("email already exists")
		}
		return nil, err
	}

	createdUser.Password = "" // Clear password before returning
	log.Info().Str("user_id", createdUser.ID.Hex()).Str("email", createdUser.Email).Msg("User registered successfully")

	if count, err := s.GetTotalUsers(ctx); err == nil {
		s.totalUsersGauge.Set(float64(count))
	}
	return createdUser, nil
}

func (s *userService) LoginUser(ctx context.Context, creds *models.Login) (string, error) {
	log.Debug().Str("email", creds.Email).Msg("Attempting user login")
	user, err := s.userRepo.FindByEmail(ctx, creds.Email)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Warn().Str("email", creds.Email).Msg("Invalid credentials during login attempt")
			return "", fmt.Errorf("invalid credentials")
		}
		log.Error().Err(err).Str("email", creds.Email).Msg("Error finding user for login")
		return "", fmt.Errorf("internal server error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		log.Warn().Str("email", creds.Email).Msg("Invalid credentials (password mismatch) during login attempt")
		return "", fmt.Errorf("invalid credentials")
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		log.Error().Err(err).Str("user_id", user.ID.Hex()).Msg("Could not generate token for user")
		return "", fmt.Errorf("could not generate token")
	}

	log.Info().Str("user_id", user.ID.Hex()).Msg("User logged in successfully")
	return token, nil
}

func (s *userService) GetUserProfile(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	log.Debug().Str("userID", userID.Hex()).Msg("Attempting to retrieve user profile")
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Warn().Str("user_id", userID.Hex()).Msg("User not found for GetMyProfile")
			return nil, fmt.Errorf("user not found")
		}
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to fetch user profile")
		return nil, fmt.Errorf("failed to fetch user profile")
	}

	user.Password = "" // Clear password before returning
	log.Info().Str("user_id", userID.Hex()).Msg("User profile retrieved successfully")
	return user, nil
}

func (s *userService) UpdateUserProfile(ctx context.Context, userID primitive.ObjectID, updatePayload *models.UserProfileUpdate) (*models.User, error) {
	log.Debug().Str("userID", userID.Hex()).Interface("updatePayload", updatePayload).Msg("Attempting to update user profile")
	updateFields := bson.M{}
	if updatePayload.Username != "" {
		updateFields["username"] = updatePayload.Username
	}
	if updatePayload.Email != nil {
		currentUser, err := s.userRepo.FindByID(ctx, userID)
		if err != nil {
			log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to verify current user data for profile update")
			return nil, fmt.Errorf("failed to verify current user data: %w", err)
		}

		if currentUser.Email != *updatePayload.Email {
			existingUser, err := s.userRepo.FindByEmail(ctx, *updatePayload.Email)
			if err == nil && existingUser != nil {
				log.Warn().Str("email", *updatePayload.Email).Msg("Email already in use by another account during profile update")
				return nil, fmt.Errorf("email already in use by another account")
			} else if err != mongo.ErrNoDocuments {
				log.Error().Err(err).Str("email", *updatePayload.Email).Msg("Failed to check email availability during profile update")
				return nil, fmt.Errorf("failed to check email availability: %w", err)
			}
		}
		updateFields["email"] = *updatePayload.Email
	}
	if updatePayload.Password != nil && *updatePayload.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updatePayload.Password), 8)
		if err != nil {
			log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Failed to hash new password for profile update")
			return nil, fmt.Errorf("failed to hash new password: %w", err)
		}
		updateFields["password"] = string(hashedPassword)
	}

	if len(updateFields) == 0 {
		log.Warn().Str("userID", userID.Hex()).Msg("No valid fields provided for user profile update")
		return nil, fmt.Errorf("no valid fields provided for update")
	}

	result, err := s.userRepo.Update(ctx, userID, updateFields)
	if err != nil {
		return nil, err
	}

	if result.MatchedCount == 0 {
		log.Warn().Str("user_id", userID.Hex()).Msg("User not found or not authorized to update profile")
		return nil, fmt.Errorf("user not found or not authorized to update")
	}

	updatedUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.Hex()).Msg("Error fetching updated user profile")
		return nil, fmt.Errorf("failed to retrieve updated user profile")
	}
	updatedUser.Password = ""

	log.Info().Str("user_id", userID.Hex()).Msg("User profile updated successfully")
	return updatedUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, userID primitive.ObjectID) error {
	log.Debug().Str("userID", userID.Hex()).Msg("Attempting to delete user account")
	result, err := s.userRepo.Delete(ctx, userID)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		log.Warn().Str("user_id", userID.Hex()).Msg("User account not found or not authorized to delete")
		return fmt.Errorf("user account not found or not authorized to delete")
	}

	log.Info().Str("user_id", userID.Hex()).Msg("User account deleted successfully")

	if count, err := s.GetTotalUsers(ctx); err == nil {
		s.totalUsersGauge.Set(float64(count))
	}
	return nil
}
