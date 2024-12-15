package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	os.Exit(main_helper())
}

func main_helper() int {
	filepathRoot := "."
	port := "8080"

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Add the readiness endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Custom FileServer to handle /app/ path
	mux.Handle("/app/", http.StripPrefix("/app", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(filepathRoot, r.URL.Path)

		// Serve the file
		http.ServeFile(w, r, filePath)
	})))

	// Create a new http.Server struct
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Start the server
	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Server failed to start: %v\n", err)
		return 1
	}

	// No errors, return 0
	return 0
}
