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

type CleanedResponse struct {
	CleanedBody string `json:"cleaned_body"`
}
