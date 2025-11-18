package main

import (
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	_ "github.com/joho/godotenv/autoload" // Import godotenv/autoload
	"markly/internal/server"
)

func main() {
	// Configure zerolog for better output
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	s := server.NewServer()

	done := make(chan bool, 1)

	go s.GracefulShutdown(done)

	err := s.Start()
	if err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("HTTP server error")
	}

	<-done
	log.Info().Msg("Graceful shutdown complete.")
}
