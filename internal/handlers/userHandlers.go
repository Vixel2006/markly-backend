package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"markly/internal/models"
	"markly/internal/services"
	"markly/internal/utils"
)

type UserHandler struct {
	userService services.UserService
}

func NewUserHandler(userService services.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (u *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Error().Err(err).Msg("Invalid user data input for Register")
		utils.SendJSONError(w, "Invalid user data input: "+err.Error(), http.StatusBadRequest)
		return
	}

	registeredUser, err := u.userService.RegisterUser(r.Context(), &user)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "invalid") {
			statusCode = http.StatusBadRequest
		} else if strings.Contains(err.Error(), "already exists") {
			statusCode = http.StatusConflict
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, registeredUser)
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.Login

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		log.Error().Err(err).Msg("Invalid request body for Login")
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	token, err := u.userService.LoginUser(r.Context(), &creds)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "invalid credentials") {
			statusCode = http.StatusUnauthorized
		}
		utils.RespondWithError(w, statusCode, err.Error())
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (u *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Error().Msg("User ID not found in context for GetMyProfile")
		utils.SendJSONError(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user ID format in context for GetMyProfile")
		utils.SendJSONError(w, "Invalid user ID format in context", http.StatusInternalServerError)
		return
	}

	user, err := u.userService.GetUserProfile(r.Context(), userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, user)
}

func (u *UserHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Error().Msg("User ID not found in context for UpdateMyProfile")
		utils.SendJSONError(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user ID format for UpdateMyProfile")
		utils.SendJSONError(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	var updatePayload models.UserProfileUpdate
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		log.Error().Err(err).Msg("Invalid JSON payload for UpdateMyProfile")
		utils.SendJSONError(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updatedUser, err := u.userService.UpdateUserProfile(r.Context(), userID, &updatePayload)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not authorized") {
			statusCode = http.StatusNotFound
		} else if strings.Contains(err.Error(), "email already in use") {
			statusCode = http.StatusConflict
		} else if strings.Contains(err.Error(), "no valid fields provided") {
			statusCode = http.StatusBadRequest
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, updatedUser)
}

func (u *UserHandler) DeleteMyProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		log.Error().Msg("User ID not found in context for DeleteMyProfile")
		utils.SendJSONError(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		log.Error().Err(err).Str("user_id_str", userIDStr).Msg("Invalid user ID format for DeleteMyProfile")
		utils.SendJSONError(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	err = u.userService.DeleteUser(r.Context(), userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "not authorized") {
			statusCode = http.StatusNotFound
		}
		utils.SendJSONError(w, err.Error(), statusCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
