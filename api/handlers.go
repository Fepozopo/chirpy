package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// HandleMetrics responds with a simple HTML page displaying the current value of the
// file server hit counter.
func (cfg *ApiConfig) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	hits := cfg.fileserverHits.Load()
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html>\n<body>\n<h1>Welcome, Chirpy Admin</h1>\n<p>Chirpy has been visited %d times!</p>\n</body>\n</html>\n", hits)
}

// handleReset resets the file server hit counter to zero and responds with a plain text message
// indicating that the hits have been reset.
func (cfg *ApiConfig) HandleReset(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0\n"))
}

// HandleValidChirp validates and processes a chirp request. It parses the JSON
// body into a ChirpRequest struct, checks if the chirp exceeds 140 characters,
// and replaces any profane words with "****". If the request body is invalid or
// the chirp is too long, it responds with a 400 Bad Request and an error message.
// Otherwise, it responds with a 200 OK and a CleanedResponse containing the
// sanitized chirp text.
func HandleValidChirp(w http.ResponseWriter, r *http.Request) {
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

	// Respond with 200 OK and a valid response if the chirp is within the allowed limit
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(CleanedResponse{CleanedBody: cleanedBody})
}
