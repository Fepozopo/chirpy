package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
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

// HandleCreateChirp creates a new chirp from the request body and saves it in the
// database. It also performs some basic validation on the request body, such as
// checking that the chirp is not too long and that it doesn't contain certain
// profane words. If the chirp is valid, it responds with a 201 status code and
// the full chirp resource. If the chirp is invalid, it responds with a 400 status
// code and an error message. If the server encounters an internal error when
// saving the chirp, it responds with a 500 status code and an error message.
func (cfg *ApiConfig) HandleCreateChirp(w http.ResponseWriter, r *http.Request) {
	// Parse the JSON body of the request into a CreateChirpRequest struct
	var createChirpRequest CreateChirpRequest
	if err := json.NewDecoder(r.Body).Decode(&createChirpRequest); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Invalid request body"})
		return
	}

	// Check if the chirp exceeds the 140 character limit
	if len(createChirpRequest.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Chirp is too long"})
		return
	}

	// Define the list of profane words to replace
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	// Replace profane words with **** in a case-insensitive manner
	cleanedBody := createChirpRequest.Body
	caser := cases.Title(language.English)
	for _, word := range profaneWords {
		cleanedBody = strings.ReplaceAll(cleanedBody, word, "****")
		cleanedBody = strings.ReplaceAll(cleanedBody, caser.String(word), "****")
		cleanedBody = strings.ReplaceAll(cleanedBody, strings.ToUpper(word), "****")
	}
	createChirpRequest.Body = cleanedBody

	// If the chirp is valid, save it in the database
	chirp, err := cfg.DbQueries.CreateChirp(r.Context(), database.CreateChirpParams(createChirpRequest))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to create chirp"})
		return
	}

	// Map the chirp struct to a MappedChirp struct to control the JSON keys
	mappedChirp := MappedChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	// If creating the record goes well, respond with a 201 status code and the full chirp resource
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mappedChirp)
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

	// Map user to the MappedUser strut in order to control the JSON keys
	mappedUser := MappedUser{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	// Respond with 200 OK and a valid response if the user was created successfully
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(mappedUser)
}

// HandleGetAllChirps retrieves all chirps from the database and returns them
// as a JSON array in the response. It maps the database chirp records to the
// MappedChirp struct to ensure consistent JSON keys. If the operation is
// successful, it responds with a 200 OK status and a pretty-printed JSON
// array of chirps. If there is an error accessing the database, it responds
// with a 500 status and an error message.
func (cfg *ApiConfig) HandleGetAllChirps(w http.ResponseWriter, r *http.Request) {
	// Get all chirps from the database
	chirps, err := cfg.DbQueries.GetAllChirps(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to get all chirps"})
		return
	}

	// Map the chirps to the MappedChirp struct to control the JSON keys
	var mappedChirps []MappedChirp
	for _, chirp := range chirps {
		mappedChirp := MappedChirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		mappedChirps = append(mappedChirps, mappedChirp)
	}

	// Respond with 200 OK and a valid response if successful
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// Create a new JSON encoder and configure it for pretty printing
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ") // Set indent with two spaces
	encoder.Encode(mappedChirps)
}

func (cfg *ApiConfig) HandleGetChirp(w http.ResponseWriter, r *http.Request) {
	// Get the chirp ID from the path parameter and convert it to a UUID object
	pathParameter := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(pathParameter)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to convert path parameter to UUID"})
		return
	}

	// Get the requested chirp from the database
	chirp, err := cfg.DbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Failed to find chirp with ID: " + pathParameter})
		return
	}

	// Map the chirp struct to a MappedChirp struct to control the JSON keys
	mappedChirp := MappedChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	// If the chirp is found, respond with a 200 OK code and the found chirp
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(mappedChirp)
}
