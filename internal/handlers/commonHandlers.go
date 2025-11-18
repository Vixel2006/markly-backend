package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"markly/internal/database"
)

type CommonHandler struct {
	db database.Service
}

func NewCommonHandler(db database.Service) *CommonHandler {
	return &CommonHandler{db: db}
}

func (h *CommonHandler) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatal().Err(err).Msg("Error marshalling JSON response for HelloWorldHandler")
	}

	_, _ = w.Write(jsonResp)
}

func (h *CommonHandler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(h.db.Health())

	if err != nil {
		log.Fatal().Err(err).Msg("Error marshalling JSON response for HealthHandler")
	}

	_, _ = w.Write(jsonResp)
}