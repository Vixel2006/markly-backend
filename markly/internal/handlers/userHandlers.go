package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"markly/internal/database"
	"markly/internal/models"
	"markly/internal/utils"
)

type UserHandler struct {
	db database.Service
}

func NewUserHandler(db database.Service) *UserHandler {
	return &UserHandler{db: db}
}

func (u *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var user models.User

	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		utils.SendJSONError(w, "Invalid user data input: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Email == "" || user.Password == "" {
		utils.SendJSONError(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		utils.SendJSONError(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	user.ID = primitive.NewObjectID()

	collection := u.db.Client().Database("markly").Collection("users")

	if err := utils.CreateUniqueIndex(collection, bson.M{"email": 1}, "Email"); err != nil {
		if strings.Contains(err.Error(), "already exists") {
			utils.SendJSONError(w, err.Error(), http.StatusConflict)
		} else {
			log.Printf("Error creating unique email index: %v", err)
			utils.SendJSONError(w, "Failed to set up database index", http.StatusInternalServerError)
		}
		return
	}

	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			utils.SendJSONError(w, "Email already exists", http.StatusConflict)
			return
		}
		log.Printf("Failed to insert user into database: %v", err)
		utils.SendJSONError(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user.Password = ""
	utils.RespondWithJSON(w, http.StatusCreated, user)
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.Login

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	collection := u.db.Client().Database("markly").Collection("users")

	var user models.User
	err := collection.FindOne(context.Background(), bson.M{"email": creds.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}
		log.Printf("Error finding user for login: %v", err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		utils.RespondWithError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("Could not generate token for user %s: %v", user.ID.Hex(), err)
		utils.RespondWithError(w, http.StatusInternalServerError, "Could not generate token")
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (u *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		utils.SendJSONError(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.SendJSONError(w, "Invalid user ID format in context", http.StatusInternalServerError)
		return
	}

	var user models.User
	collection := u.db.Client().Database("markly").Collection("users")

	filter := bson.M{"_id": userID}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.SendJSONError(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to fetch user profile for %s: %v", userID.Hex(), err)
		utils.SendJSONError(w, "Failed to fetch user profile", http.StatusInternalServerError)
		return
	}

	user.Password = ""

	utils.RespondWithJSON(w, http.StatusOK, user)
}

func (u *UserHandler) buildUserProfileUpdateFields(updatePayload models.UserProfileUpdate, userID primitive.ObjectID) (bson.M, error) {
	updateFields := bson.M{}
	if updatePayload.Username != "" {
		updateFields["username"] = updatePayload.Username
	}
	if updatePayload.Email != nil {
		var currentUser models.User
		err := u.db.Client().Database("markly").Collection("users").FindOne(context.Background(), bson.M{"_id": userID}).Decode(&currentUser)
		if err != nil {
			return nil, fmt.Errorf("failed to verify current user data: %w", err)
		}

		if currentUser.Email != *updatePayload.Email {
			var existingUser models.User
			err := u.db.Client().Database("markly").Collection("users").
				FindOne(context.Background(), bson.M{"email": *updatePayload.Email}).Decode(&existingUser)
			if err == nil {
				return nil, fmt.Errorf("email already in use by another account")
			} else if err != mongo.ErrNoDocuments {
				return nil, fmt.Errorf("failed to check email availability: %w", err)
			}
		}
		updateFields["email"] = *updatePayload.Email
	}
	if updatePayload.Password != nil && *updatePayload.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updatePayload.Password), 8)
		if err != nil {
			return nil, fmt.Errorf("failed to hash new password: %w", err)
		}
		updateFields["password"] = string(hashedPassword)
	}
	return updateFields, nil
}

func (u *UserHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		utils.SendJSONError(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.SendJSONError(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	var updatePayload models.UserProfileUpdate
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		utils.SendJSONError(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updateFields, err := u.buildUserProfileUpdateFields(updatePayload, userID)
	if err != nil {
		utils.SendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(updateFields) == 0 {
		utils.SendJSONError(w, "No valid fields provided for update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": updateFields}

	collection := u.db.Client().Database("markly").Collection("users")
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error updating user profile %s: %v", userID.Hex(), err)
		utils.SendJSONError(w, "Failed to update user profile", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		utils.SendJSONError(w, "User not found or not authorized to update", http.StatusNotFound)
		return
	}

	var updatedUser models.User
	err = collection.FindOne(context.Background(), filter).Decode(&updatedUser)
	if err != nil {
		log.Printf("Error fetching updated user profile %s: %v", userID.Hex(), err)
		utils.SendJSONError(w, "Failed to retrieve updated user profile", http.StatusInternalServerError)
		return
	}
	updatedUser.Password = ""

	utils.RespondWithJSON(w, http.StatusOK, updatedUser)
}

func (u *UserHandler) DeleteMyProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		utils.SendJSONError(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		utils.SendJSONError(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	collection := u.db.Client().Database("markly").Collection("users")

	filter := bson.M{"_id": userID}
	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Error deleting user account %s: %v", userID.Hex(), err)
		utils.SendJSONError(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		utils.SendJSONError(w, "User account not found or not authorized to delete", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
