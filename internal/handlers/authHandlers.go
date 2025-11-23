package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
	"github.com/rs/zerolog/log"

	"markly/internal/services"
)

type AuthHandler struct {
	authService services.AuthService
}

func NewAuthHandler(AuthService services.AuthService) *AuthHandler {
	return &AuthHandler{authService: AuthService}
}

func (a *AuthHandler) ProviderAuth(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	provider := vars["provider"]

	if provider == "" {
		log.Error().Msg("Provider not specified in URL")
		http.Error(w, "Provider not specified", http.StatusBadRequest)
		return
	}

	log.Info().Str("provider", provider).Msg("Initiating authentication with provider")

	gothic.BeginAuthHandler(w, r)
}

func (a *AuthHandler) ProviderCallback(w http.ResponseWriter, r *http.Request) {
	log.Info().Msg("Provider callback initiated")

	PUser, err := gothic.CompleteUserAuth(w, r)

	if err != nil {
		log.Error().Err(err).Msg("Error completing user authentication")
		http.Redirect(w, r, "/api/auth/error", http.StatusTemporaryRedirect)
		return
	}

	log.Info().Str("email", PUser.Email).Msg("User authenticated with provider, attempting to handle login")
	token, err := a.authService.HandleLogin(r.Context(), PUser)

	if err != nil {
		log.Error().Err(err).Msg("Error handling login after provider authentication")
		http.Redirect(w, r, "/api/auth/error", http.StatusTemporaryRedirect)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
	})
	log.Info().Str("email", PUser.Email).Msg("JWT cookie set successfully")

	http.Redirect(w, r, "/api/auth/success", http.StatusTemporaryRedirect)
}

func (a *AuthHandler) AuthSuccess(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Authentication successful! Redirecting..."))
}

func (a *AuthHandler) AuthError(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Authentication failed. Please try again.", http.StatusBadRequest)
}
