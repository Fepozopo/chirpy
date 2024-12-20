package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/Fepozopo/chirpy/internal/database"
)

// HandleHealthz is a simple health-check endpoint that responds with a 200 OK and the
// plain text string "OK" to indicate that the server is running.
func (cfg *ApiConfig) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

// HandleMetrics responds with a simple HTML page displaying the current value of the
// file server hit counter.
func (cfg *ApiConfig) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	hits := cfg.fileserverHits.Load()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>\n", hits)
}

// HandleReset is a special admin-only endpoint that can only be accessed in a local
// development environment. It resets the server's hits counter to 0 and deletes all
// users in the database. It responds with a 200 OK and a plaintext message
// indicating that the hits counter has been reset to 0.
func (cfg *ApiConfig) HandleReset(w http.ResponseWriter, r *http.Request) {
	godotenv.Load()
	platform := os.Getenv("PLATFORM")
	apiCfg := ApiConfig{
		Platform: platform,
	}
	// Ensure this endpoint can only be accessed in a local development environment
	if apiCfg.Platform != "dev" {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("403 Forbidden\n"))
		return
	}

	// Delete all users in the database
	err := cfg.DbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed delete users"})
		return
	}

	// Reset the server hits count to 0
	cfg.fileserverHits.Store(0)

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0 and all users deleted\n"))
}

func (cfg *ApiConfig) HandleChirps(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON body of the request into a ChirpRequest struct
	var chirpRequest ChirpRequest
	if err := json.NewDecoder(r.Body).Decode(&chirpRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Check if the chirp exceeds the 140 character limit
	if len(chirpRequest.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Chirp is too long"})
		return
	}

	// Define the list of profane words to replace
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	// Replace profane words with **** in a case-insensitive manner
	cleanedBody := chirpRequest.Body
	caser := cases.Title(language.English)
	for _, word := range profaneWords {
		cleanedBody = strings.ReplaceAll(cleanedBody, word, "****")
		cleanedBody = strings.ReplaceAll(cleanedBody, caser.String(word), "****")
		cleanedBody = strings.ReplaceAll(cleanedBody, strings.ToUpper(word), "****")
	}
	chirpRequest.Body = cleanedBody

	// If the chirp is valid, save it in the database
	chirp, err := cfg.DbQueries.CreateChirp(r.Context(), database.CreateChirpParams(chirpRequest))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create chirp"})
		return
	}

	// Map the chirp struct to a MapChirp struct to control the JSON keys
	mapChirp := MapChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	// If creating the record goes well, respond with a 201 status code and the full chirp resource
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapChirp)
}

// HandleCreateUser creates a new user from the email address in the request body
// and returns the user's ID, email, and timestamps in the response body.
func (cfg *ApiConfig) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON body of the request into a CreateUserRequest struct
	var createUserRequest CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&createUserRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Create a new user in the database
	user, err := cfg.DbQueries.CreateUser(r.Context(), createUserRequest.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create user"})
		return
	}

	// Map user to the User strut in order to control the JSON keys
	mapUser := MapUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	// Respond with 200 OK and a valid response if the user was created successfully
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mapUser)
}
