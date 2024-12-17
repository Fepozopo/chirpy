package api

import (
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
