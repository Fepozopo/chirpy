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
	Email    string `json:"email"`
	Password string `json:"password"`
}

type NewJWT struct {
	Token string `json:"token"`
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type StripeEvent struct {
	Event string `json:"event"`
	Data  struct {
		UserID uuid.UUID `json:"user_id"`
	} `json:"data"`
}

type MappedUser struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	IsChirpyRed  bool      `json:"is_chirpy_red"`
	Token        string    `json:"token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
}

type MappedChirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}
