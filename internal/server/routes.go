package server

import (
	"net/http"

	"github.com/gorilla/mux"

	"markly/internal/handlers"
	"markly/internal/middlewares"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()

	r.Use(middlewares.CorsMiddleware)
	r.Use(middlewares.RateLimit)

	ch := handlers.NewCommonHandler(s.db)
	r.HandleFunc("/", ch.HelloWorldHandler)
	r.HandleFunc("/health", ch.HealthHandler)

	s.registerBookmarkRoutes(r)
	s.registerAuthRoutes(r)
	s.registerTagRoutes(r)
	s.registerCollectionRoutes(r)
	s.registerCategoryRoutes(r)
	s.registerAgentRoutes(r)

	return r
}

func (s *Server) registerBookmarkRoutes(r *mux.Router) {
	bh := handlers.NewBookmarksHandler(s.bookmarkService)

	r.Handle("/api/bookmarks", middlewares.AuthMiddleware(http.HandlerFunc(bh.GetBookmarks))).Methods("GET", "OPTIONS")
	r.Handle("/api/bookmarks", middlewares.AuthMiddleware(http.HandlerFunc(bh.AddBookmark))).Methods("POST", "OPTIONS")
	r.Handle("/api/bookmarks/{id}", middlewares.AuthMiddleware(http.HandlerFunc(bh.GetBookmarkByID))).Methods("GET", "OPTIONS")
	r.Handle("/api/bookmarks/{id}", middlewares.AuthMiddleware(http.HandlerFunc(bh.DeleteBookmark))).Methods("DELETE", "OPTIONS")
	r.Handle("/api/bookmarks/{id}", middlewares.AuthMiddleware(http.HandlerFunc(bh.UpdateBookmark))).Methods("PUT", "OPTIONS")
}

func (s *Server) registerAuthRoutes(r *mux.Router) {
	uh := handlers.NewUserHandler(s.userService)
	ah := handlers.NewAuthHandler(s.authService, s.otpService)

	r.HandleFunc("/api/auth/register", uh.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/login", uh.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/forgot-password", ah.ForgotPasswordHandler).Methods("POST", "OPTIONS")
	r.Handle("/api/me", middlewares.AuthMiddleware(http.HandlerFunc(uh.GetMyProfile))).Methods("GET", "OPTIONS")
	r.Handle("/api/me", middlewares.AuthMiddleware(http.HandlerFunc(uh.UpdateMyProfile))).Methods("PATCH", "PUT", "OPTIONS")
	r.Handle("/api/me", middlewares.AuthMiddleware(http.HandlerFunc(uh.DeleteMyProfile))).Methods("DELETE", "OPTIONS")

	r.HandleFunc("/api/auth/{provider}", ah.ProviderAuth).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/auth/{provider}/callback", ah.ProviderCallback).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/auth/success", ah.AuthSuccess).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/auth/error", ah.AuthError).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/auth/reset-password", ah.ResetPasswordHandler).Methods("POST", "OPTIONS")
}

func (s *Server) registerCategoryRoutes(r *mux.Router) {
	ch := handlers.NewCategoryHandler(s.categoryService)
	r.Handle("/api/categories", middlewares.AuthMiddleware(http.HandlerFunc(ch.AddCategory))).Methods("POST", "OPTIONS")
	r.Handle("/api/categories", middlewares.AuthMiddleware(http.HandlerFunc(ch.GetCategories))).Methods("GET", "OPTIONS")
	r.Handle("/api/categories/{id}", middlewares.AuthMiddleware(http.HandlerFunc(ch.GetCategoryByID))).Methods("GET", "OPTIONS")
	r.Handle("/api/categories/{id}", middlewares.AuthMiddleware(http.HandlerFunc(ch.DeleteCategory))).Methods("DELETE", "OPTIONS")
	r.Handle("/api/categories/{id}", middlewares.AuthMiddleware(http.HandlerFunc(ch.UpdateCategory))).Methods("PUT", "OPTIONS")
}

func (s *Server) registerCollectionRoutes(r *mux.Router) {
	clh := handlers.NewCollectionHandler(s.collectionService)
	r.Handle("/api/collections", middlewares.AuthMiddleware(http.HandlerFunc(clh.AddCollection))).Methods("POST", "OPTIONS")
	r.Handle("/api/collections", middlewares.AuthMiddleware(http.HandlerFunc(clh.GetCollections))).Methods("GET", "OPTIONS")
	r.Handle("/api/collections/{id}", middlewares.AuthMiddleware(http.HandlerFunc(clh.GetCollection))).Methods("GET", "OPTIONS")
	r.Handle("/api/collections/{id}", middlewares.AuthMiddleware(http.HandlerFunc(clh.DeleteCollection))).Methods("DELETE", "OPTIONS")
	r.Handle("/api/collections/{id}", middlewares.AuthMiddleware(http.HandlerFunc(clh.UpdateCollection))).Methods("PUT", "OPTIONS")
}

func (s *Server) registerTagRoutes(r *mux.Router) {
	th := handlers.NewTagHandler(s.tagService)
	r.Handle("/api/tags", middlewares.AuthMiddleware(http.HandlerFunc(th.AddTag))).Methods("POST", "OPTIONS")
	r.Handle("/api/tags", middlewares.AuthMiddleware(http.HandlerFunc(th.GetTagsByID))).Methods("GET", "OPTIONS")
	r.Handle("/api/tags/user", middlewares.AuthMiddleware(http.HandlerFunc(th.GetUserTags))).Methods("GET", "OPTIONS")
	r.Handle("/api/tags/{id}", middlewares.AuthMiddleware(http.HandlerFunc(th.DeleteTag))).Methods("DELETE", "OPTIONS")
	r.Handle("/api/tags/{id}", middlewares.AuthMiddleware(http.HandlerFunc(th.UpdateTag))).Methods("PUT", "OPTIONS")
}

func (s *Server) registerAgentRoutes(r *mux.Router) {
	ah := handlers.NewAgentHandler(s.agentService)
	r.Handle("/api/agent/summarize/{id}", middlewares.AuthMiddleware(http.HandlerFunc(ah.GenerateSummary))).Methods("POST", "OPTIONS")
	r.Handle("/api/agent/summarize-url", middlewares.AuthMiddleware(http.HandlerFunc(ah.SummarizeURL))).Methods("POST", "OPTIONS")
	r.Handle("/api/agent/suggestions", middlewares.AuthMiddleware(http.HandlerFunc(ah.GenerateAISuggestions))).Methods("GET", "OPTIONS")
}
