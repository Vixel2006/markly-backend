package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	_ "github.com/joho/godotenv/autoload"

	"markly/internal/database"
	"markly/internal/repositories"
	"markly/internal/services"
)

type Server struct {
	port              int
	httpServer        *http.Server
	db                database.Service
	userService       services.UserService
	bookmarkService   services.BookmarkService
	categoryService   services.CategoryService
	collectionService services.CollectionService
	tagService        services.TagService
	agentService      *services.AgentService
	authService       services.AuthService
}

func NewServer() *Server {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal().Err(err).Str("port", portStr).Msgf("Invalid PORT environment variable. Using default 8080.")
		port = 8080
	}

	db := database.New()

	userRepo := repositories.NewUserRepository(db)
	bookmarkRepo := repositories.NewBookmarkRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	collectionRepo := repositories.NewCollectionRepository(db)
	tagRepo := repositories.NewTagRepository(db)

	authService := services.NewAuthService(userRepo)

	s := &Server{
		port:              port,
		db:                db,
		userService:       services.NewUserService(userRepo),
		bookmarkService:   services.NewBookmarkService(bookmarkRepo, db),
		categoryService:   services.NewCategoryService(categoryRepo),
		collectionService: services.NewCollectionService(collectionRepo),
		tagService:        services.NewTagService(tagRepo),
		agentService:      services.NewAgentService(bookmarkRepo, categoryRepo, collectionRepo, tagRepo),
		authService:       authService,
	}

	services.InitializeGoth()

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return s
}

func (s *Server) Start() error {
	log.Info().Int("port", s.port).Msg("Starting server")
	return s.httpServer.ListenAndServe()
}

func (s *Server) GracefulShutdown(done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Info().Msg("Shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown with error")
	}

	log.Info().Msg("Server exiting")
	done <- true
}
