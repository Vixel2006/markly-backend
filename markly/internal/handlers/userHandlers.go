package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
		http.Error(w, "Invalid user data input: "+err.Error(), http.StatusBadRequest)
		return
	}

	if user.Username == "" || user.Email == "" || user.Password == "" {
		http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	user.ID = primitive.NewObjectID()

	collection := u.db.Client().Database("markly").Collection("users")

	indexModel := mongo.IndexModel{
		Keys:    bson.M{"email": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err = collection.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		// Check if it's a duplicate key error using strings.Contains
		if !strings.Contains(err.Error(), "E11000 duplicate key error collection") {
			log.Printf("Error creating unique email index: %v", err)
			http.Error(w, "Failed to set up database index", http.StatusInternalServerError)
			return
		}
	}

	_, err = collection.InsertOne(context.Background(), user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			http.Error(w, "Email already exists", http.StatusConflict)
			return
		}
		log.Printf("Failed to insert user into database: %v", err)
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	user.Password = ""
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func (u *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var creds models.Login

	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	collection := u.db.Client().Database("markly").Collection("users")

	var user models.User
	err := collection.FindOne(context.Background(), bson.M{"email": creds.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		log.Printf("Error finding user for login: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		log.Printf("Could not generate token for user %s: %v", user.ID.Hex(), err)
		http.Error(w, "Could not generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (u *UserHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusInternalServerError)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format in context", http.StatusInternalServerError)
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
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		log.Printf("Failed to fetch user profile for %s: %v", userIDStr, err)
		http.Error(w, "Failed to fetch user profile", http.StatusInternalServerError)
		return
	}

	user.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (u *UserHandler) UpdateMyProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	var updatePayload models.UserProfileUpdate
	if err := json.NewDecoder(r.Body).Decode(&updatePayload); err != nil {
		http.Error(w, "Invalid JSON payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	updateFields := bson.M{}
	if updatePayload.Username != "" {
		updateFields["username"] = updatePayload.Username
	}
	if updatePayload.Email != nil {
		var currentUser models.User
		err := u.db.Client().Database("markly").Collection("users").FindOne(context.Background(), bson.M{"_id": userID}).Decode(&currentUser)
		if err != nil {
			log.Printf("Error fetching current user for email update check: %v", err)
			http.Error(w, "Failed to verify current user data", http.StatusInternalServerError)
			return
		}

		if currentUser.Email != *updatePayload.Email {
			var existingUser models.User
			err := u.db.Client().Database("markly").Collection("users").
				FindOne(context.Background(), bson.M{"email": *updatePayload.Email}).Decode(&existingUser)
			if err == nil {
				http.Error(w, "Email already in use by another account.", http.StatusConflict)
				return
			} else if err != mongo.ErrNoDocuments {
				log.Printf("Error checking for duplicate email: %v", err)
				http.Error(w, "Failed to check email availability.", http.StatusInternalServerError)
				return
			}
		}
		updateFields["email"] = *updatePayload.Email
	}
	if updatePayload.Password != nil && *updatePayload.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*updatePayload.Password), 8)
		if err != nil {
			http.Error(w, "Failed to hash new password", http.StatusInternalServerError)
			return
		}
		updateFields["password"] = string(hashedPassword)
	}

	if len(updateFields) == 0 {
		http.Error(w, "No valid fields provided for update", http.StatusBadRequest)
		return
	}

	filter := bson.M{"_id": userID}
	update := bson.M{"$set": updateFields}

	collection := u.db.Client().Database("markly").Collection("users")
	result, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		log.Printf("Error updating user profile %s: %v", userIDStr, err)
		http.Error(w, "Failed to update user profile", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "User not found or not authorized to update", http.StatusNotFound)
		return
	}

	var updatedUser models.User
	err = collection.FindOne(context.Background(), filter).Decode(&updatedUser)
	if err != nil {
		log.Printf("Error fetching updated user profile %s: %v", userIDStr, err)
		http.Error(w, "Failed to retrieve updated user profile", http.StatusInternalServerError)
		return
	}
	updatedUser.Password = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

func (u *UserHandler) DeleteMyProfile(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		http.Error(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID format", http.StatusUnauthorized)
		return
	}

	collection := u.db.Client().Database("markly").Collection("users")

	filter := bson.M{"_id": userID}
	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		log.Printf("Error deleting user account %s: %v", userIDStr, err)
		http.Error(w, "Failed to delete account", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		http.Error(w, "User account not found or not authorized to delete", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
