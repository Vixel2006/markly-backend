package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/markbates/goth/gothic"
	"github.com/rs/zerolog/log"

	"markly/internal/services"
	"markly/internal/utils"
)

type AuthHandler struct {
	authService services.AuthService
	otpService  services.OTPService
}

func NewAuthHandler(AuthService services.AuthService, otpService services.OTPService) *AuthHandler {
	return &AuthHandler{authService: AuthService, otpService: otpService}
}

type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

func (a *AuthHandler) ForgotPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Email == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Email is required")
		return
	}

	_, err := a.otpService.GenerateOTPForgotPassword(r.Context(), req.Email)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to generate and send OTP for password reset")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to send OTP for password reset")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Password reset OTP sent successfully"})
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

type ResetPasswordRequest struct {
	Email       string `json:"email"`
	OTP         string `json:"otp"`
	NewPassword string `json:"new_password"`
}

func (a *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if req.Email == "" || req.OTP == "" || req.NewPassword == "" {
		utils.RespondWithError(w, http.StatusBadRequest, "Email, OTP, and new password are required")
		return
	}

	// Verify OTP
	err := a.otpService.VerifyOTP(r.Context(), req.Email, req.OTP)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("OTP verification failed")
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid or expired OTP")
		return
	}

	// Reset password
	err = a.authService.ResetPassword(r.Context(), req.Email, req.NewPassword)
	if err != nil {
		log.Error().Err(err).Str("email", req.Email).Msg("Failed to reset password")
		utils.RespondWithError(w, http.StatusInternalServerError, "Failed to reset password")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Password reset successfully"})
}
