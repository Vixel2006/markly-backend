package services

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/google"
	"github.com/rs/zerolog/log"

	"markly/internal/models"
	"markly/internal/repositories"
	"markly/internal/utils"
)

const (
	key    = "random string"
	MaxAge = 86400 * 30
	IsProd = false
)

type AuthService interface {
	HandleLogin(ctx context.Context, u goth.User) (string, error)
}

type authService struct {
	userRepo repositories.UserRepository
}

func NewAuthService(UserRepo repositories.UserRepository) *authService {
	return &authService{userRepo: UserRepo}
}

func InitializeGoth() {
	// TODO: Ensure InitializeGoth() is called once and early in the application lifecycle.
	google_client_id := os.Getenv("GOOGLE_CLIENT_ID")
	google_client_secret := os.Getenv("GOOGLE_CLIENT_SECRET")
	facebook_client_id := os.Getenv("FACEBOOK_CLIENT_ID")
	facebook_client_secret := os.Getenv("FACEBOOK_CLIENT_SECRET")

	sessionKey := os.Getenv("SESSION_KEY")

	store := sessions.NewCookieStore([]byte(sessionKey))
	store.MaxAge(MaxAge)

	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = IsProd
	store.Options.SameSite = http.SameSiteLaxMode

	gothic.Store = store

	goth.UseProviders(
		google.New(google_client_id, google_client_secret, "http://localhost:8080/api/auth/google/callback"),
		facebook.New(facebook_client_id, facebook_client_secret, "http://localhost:8080/api/auth/microsoft/callback"),
	)
	log.Info().Msg("Goth providers initialized")
}

func (a *authService) HandleLogin(ctx context.Context, u goth.User) (string, error) {
	log.Info().Str("email", u.Email).Msg("Attempting to handle login for user")
	if u.Email == "" {
		log.Error().Msg("Missing email in Goth user data")
		return "", errors.New("missing Email")
	}

	user, err := a.userRepo.FindByEmail(ctx, u.Email)

	if err != nil {
		log.Error().Err(err).Str("email", u.Email).Msg("Error finding user by email")
		return "", errors.New("error finding user by email")
	}

	if user == nil {
		log.Info().Str("email", u.Email).Msg("User not found, creating new user")
		newUser := &models.User{
			Email:    u.Email,
			Username: u.NickName,
		}
		if _, err := a.userRepo.Create(ctx, newUser); err != nil {
			log.Error().Err(err).Str("email", u.Email).Msg("Error creating new user")
			return "", errors.New("error creating user")
		}
		user = newUser
		log.Info().Str("email", u.Email).Str("userID", user.ID.Hex()).Msg("New user created successfully")
	} else {
		log.Info().Str("email", u.Email).Str("userID", user.ID.Hex()).Msg("User found in database")
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		log.Error().Err(err).Str("userID", user.ID.Hex()).Msg("Error generating JWT for user")
		return "", errors.New("error generating JWT")
	}
	log.Info().Str("userID", user.ID.Hex()).Msg("JWT generated successfully")

	return token, nil
}
