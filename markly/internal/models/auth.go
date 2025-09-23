package models

// Login represents the credentials submitted for user login.
type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
