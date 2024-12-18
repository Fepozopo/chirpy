package api

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// HandleValidChirp validates the request body of a POST request to /api/validate_chirp as a
// ChirpRequest struct, and checks if the Body field of the request exceeds the 140 character
// limit. If either of these validations fail, the handler responds with a 400 Bad Request
// status, encoding an ErrorResponse struct into the response body. If the request is valid,
// the handler responds with a 200 OK status, encoding a ValidResponse struct into the
// response body.
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

	// Respond with 200 OK and a valid response if the chirp is within the allowed limit
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(ValidResponse{Valid: true})
}
