package api

import (
	"sync/atomic"
	"time"

	"github.com/Fepozopo/chirpy/internal/database"
	"github.com/google/uuid"
)

type ApiConfig struct {
	fileserverHits atomic.Int32
	DbQueries      *database.Queries
	Platform       string `env:"PLATFORM"`
	TokenSecret    string `env:"TOKEN_SECRET"`
}

type CreateChirpRequest struct {
	Body   string    `json:"body"`
	UserID uuid.UUID `json:"user_id"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type CreateUserRequest struct {
	Email          string `json:"email"`
	HashedPassword string `json:"password"`
}

type LoginUserRequest struct {
	Email            string `json:"email"`
	Password         string `json:"password"`
	ExpiresInSeconds *int   `json:"expires_in_seconds,omitempty"`
}

type MappedUser struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token,omitempty"`
}

type MappedChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
