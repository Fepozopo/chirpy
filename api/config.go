package api

import (
	"sync/atomic"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
}

type ChirpRequest struct {
	Body string `json:"body"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ValidResponse struct {
	Valid bool `json:"valid"`
}
