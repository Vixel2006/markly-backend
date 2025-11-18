package utils

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetUserIDFromContext extracts and parses the userID from the request context.
func GetUserIDFromContext(w http.ResponseWriter, r *http.Request) (primitive.ObjectID, error) {
	userIDStr, ok := r.Context().Value("userID").(string)
	if !ok {
		SendJSONError(w, "Invalid user ID", http.StatusUnauthorized)
		return primitive.NilObjectID, errors.New("invalid user ID in context")
	}

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		SendJSONError(w, "Invalid user ID format", http.StatusUnauthorized)
		return primitive.NilObjectID, errors.New("invalid user ID format in context")
	}
	return userID, nil
}

// GetObjectIDFromVars extracts and parses an ObjectID from mux.Vars.
func GetObjectIDFromVars(w http.ResponseWriter, r *http.Request, paramName string) (primitive.ObjectID, error) {
	vars := mux.Vars(r)
	idStr := vars[paramName]
	if idStr == "" {
		SendJSONError(w, "Missing ID parameter", http.StatusBadRequest)
		return primitive.NilObjectID, errors.New("missing ID parameter")
	}

	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		SendJSONError(w, "Invalid ID format", http.StatusBadRequest)
		return primitive.NilObjectID, errors.New("invalid ID format")
	}
	return objID, nil
}