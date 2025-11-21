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
	"markly/internal/services"
)

type Server struct {
	port int
	httpServer *http.Server
	db database.Service
	userService services.UserService
}

func NewServer() *Server {
	portStr := os.Getenv("PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal().Err(err).Str("port", portStr).Msgf("Invalid PORT environment variable. Using default 8080.")
		port = 8080 // Default port
	}

	s := &Server{
		port: port,
		db: database.New(),
	}
	s.userService = services.NewUserService(s.db)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.RegisterRoutes(s.userService),
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
