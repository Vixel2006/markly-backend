package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"markly/internal/handlers"
	"markly/internal/middlewares"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	// Apply CORS middleware
	r.Use(s.corsMiddleware)

	r.HandleFunc("/", s.HelloWorldHandler)

	r.HandleFunc("/health", s.healthHandler)

	// Register the User Auth and Bookmark Api
	s.registerBookmarkRoutes(r)
	s.registerAuthRoutes(r)

	return r
}

// CORS middleware
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS Headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Wildcard allows all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "false") // Credentials not allowed with wildcard origins

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) registerBookmarkRoutes(r *mux.Router) {
	bh := handlers.NewBookmarksHandler(s.db)

	r.Handle("/api/bookmarks", middlewares.AuthMiddleware(http.HandlerFunc(bh.GetBookmarks))).Methods("GET")
	r.Handle("/api/bookmarks", middlewares.AuthMiddleware(http.HandlerFunc(bh.AddBookmark))).Methods("POST")
	r.Handle("/api/bookmarks/{id}", middlewares.AuthMiddleware(http.HandlerFunc(bh.GetBookmarkByID))).Methods("GET")
	r.Handle("/api/bookmarks/{id}", middlewares.AuthMiddleware(http.HandlerFunc(bh.DeleteBookmark))).Methods("DELETE")
	r.Handle("/api/bookmarks/{id}", middlewares.AuthMiddleware(http.HandlerFunc(bh.UpdateBookmark))).Methods("PUT")
}

func (s *Server) registerAuthRoutes(r *mux.Router) {
	uh := handlers.NewUserHandler(s.db)

	r.HandleFunc("/api/auth/register", uh.Register).Methods("POST")
	r.HandleFunc("/api/auth/login", uh.Login).Methods("POST")
}

