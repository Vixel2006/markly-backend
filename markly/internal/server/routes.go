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

	r.Use(s.corsMiddleware)

	r.HandleFunc("/", s.HelloWorldHandler)

	r.HandleFunc("/health", s.healthHandler)

	s.registerBookmarkRoutes(r)
	s.registerAuthRoutes(r)
	s.registerTagRoutes(r)
	s.registerCollectionRoutes(r)
	s.registerCategoryRoutes(r)

	return r
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

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

	r.Handle("/api/bookmarks", middlewares.AuthMiddleware(http.HandlerFunc(bh.GetBookmarks))).Methods("GET", "OPTIONS")
	r.Handle("/api/bookmarks", middlewares.AuthMiddleware(http.HandlerFunc(bh.AddBookmark))).Methods("POST", "OPTIONS")
	r.Handle("/api/bookmarks/{id}", middlewares.AuthMiddleware(http.HandlerFunc(bh.GetBookmarkByID))).Methods("GET", "OPTIONS")
	r.Handle("/api/bookmarks/{id}", middlewares.AuthMiddleware(http.HandlerFunc(bh.DeleteBookmark))).Methods("DELETE", "OPTIONS")
	r.Handle("/api/bookmarks/{id}", middlewares.AuthMiddleware(http.HandlerFunc(bh.UpdateBookmark))).Methods("PUT", "OPTIONS")
}

func (s *Server) registerAuthRoutes(r *mux.Router) {
	uh := handlers.NewUserHandler(s.db)

	r.HandleFunc("/api/auth/register", uh.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/login", uh.Login).Methods("POST", "OPTIONS")
	r.Handle("/api/me", middlewares.AuthMiddleware(http.HandlerFunc(uh.GetMyProfile))).Methods("GET", "OPTIONS")
	r.Handle("/api/me", middlewares.AuthMiddleware(http.HandlerFunc(uh.UpdateMyProfile))).Methods("PATCH", "PUT", "OPTIONS")
	r.Handle("/api/me", middlewares.AuthMiddleware(http.HandlerFunc(uh.DeleteMyProfile))).Methods("DELETE", "OPTIONS")
}

func (s *Server) registerCategoryRoutes(r *mux.Router) {
	ch := handlers.NewCategoryHandler(s.db)
	r.Handle("/api/categories", middlewares.AuthMiddleware(http.HandlerFunc(ch.AddCategory))).Methods("POST", "OPTIONS")
	r.Handle("/api/categories", middlewares.AuthMiddleware(http.HandlerFunc(ch.GetCategories))).Methods("GET", "OPTIONS")
	r.Handle("/api/categories/{id}", middlewares.AuthMiddleware(http.HandlerFunc(ch.GetCategoryByID))).Methods("GET", "OPTIONS")
	r.Handle("/api/categories/{id}", middlewares.AuthMiddleware(http.HandlerFunc(ch.DeleteCategory))).Methods("DELETE", "OPTIONS")
	r.Handle("/api/categories/{id}", middlewares.AuthMiddleware(http.HandlerFunc(ch.UpdateCategory))).Methods("PUT", "OPTIONS")
}

func (s *Server) registerCollectionRoutes(r *mux.Router) {
	clh := handlers.NewCollectionHandler(s.db)
	r.Handle("/api/collections", middlewares.AuthMiddleware(http.HandlerFunc(clh.AddCollection))).Methods("POST", "OPTIONS")
	r.Handle("/api/collections", middlewares.AuthMiddleware(http.HandlerFunc(clh.GetCollections))).Methods("GET", "OPTIONS")
	r.Handle("/api/collections/{id}", middlewares.AuthMiddleware(http.HandlerFunc(clh.GetCollection))).Methods("GET", "OPTIONS")
	r.Handle("/api/collections/{id}", middlewares.AuthMiddleware(http.HandlerFunc(clh.DeleteCollection))).Methods("DELETE", "OPTIONS")
	r.Handle("/api/collections/{id}", middlewares.AuthMiddleware(http.HandlerFunc(clh.UpdateCollection))).Methods("PUT", "OPTIONS")
}

func (s *Server) registerTagRoutes(r *mux.Router) {
	th := handlers.NewTagHandler(s.db)
	r.Handle("/api/tags", middlewares.AuthMiddleware(http.HandlerFunc(th.AddTag))).Methods("POST", "OPTIONS")
	r.Handle("/api/tags", middlewares.AuthMiddleware(http.HandlerFunc(th.GetTagsByID))).Methods("GET", "OPTIONS")
	r.Handle("/api/tags/user", middlewares.AuthMiddleware(http.HandlerFunc(th.GetUserTags))).Methods("GET", "OPTIONS")
	r.Handle("/api/tags/{id}", middlewares.AuthMiddleware(http.HandlerFunc(th.DeleteTag))).Methods("DELETE", "OPTIONS")
	r.Handle("/api/tags/{id}", middlewares.AuthMiddleware(http.HandlerFunc(th.UpdateTag))).Methods("PUT", "OPTIONS")
}
