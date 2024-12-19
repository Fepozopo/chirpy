package api

import (
	"sync/atomic"

	"github.com/Fepozopo/chirpy/internal/database"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
	DbQueries      *database.Queries
}

type ChirpRequest struct {
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CleanedResponse struct {
	CleanedBody string `json:"cleaned_body"`
}
