package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	os.Exit(mainHelper())
}

func mainHelper() int {
	filepathRoot := "."
	port := "8080"

	apiCfg := &apiConfig{}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Add the readiness endpoint
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	// Custom FileServer to handle /app/ path
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.middlewareMetricsInc(fileServer))

	// Metrics endpoint
	mux.HandleFunc("GET /metrics", apiCfg.handleMetrics)

	// Reset endpoint
	mux.HandleFunc("POST /reset", apiCfg.handleReset)

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
