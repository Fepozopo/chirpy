package main

import (
	"log"
	"net/http"
	"os"

	api "github.com/Fepozopo/chirpy/api"
)

func main() {
	os.Exit(mainHelper())
}

func mainHelper() int {
	filepathRoot := "./app"
	port := "8080"

	apiCfg := &api.ApiConfig{}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Add the readiness endpoint
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	// Metrics endpoint
	mux.HandleFunc("GET /admin/metrics", apiCfg.HandleMetrics)

	// Reset endpoint
	mux.HandleFunc("POST /admin/reset", apiCfg.HandleReset)

	// Custom FileServer to handle /app/ path
	fileServer := http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(fileServer))

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
