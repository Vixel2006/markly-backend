package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/joho/godotenv/autoload" // Import godotenv/autoload
	"markly/internal/server"
)

func main() {
	s := server.NewServer()

	done := make(chan bool, 1)

	go s.GracefulShutdown(done)

	err := s.Start()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	<-done
	log.Println("Graceful shutdown complete.")
}
