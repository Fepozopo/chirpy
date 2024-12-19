package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	api "github.com/Fepozopo/chirpy/api"
	database "github.com/Fepozopo/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	os.Exit(mainHelper())
}

func mainHelper() int {
	filepathRoot := "./app"
	port := "8080"

	// Open a connection to the database
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Failed to open a connection to the database: %v\n", err)
	}
	defer db.Close()

	// Create a new Queries instance and initialize the ApiConfig struct
	dbQueries := database.New(db)
	apiCfg := &api.ApiConfig{
		DbQueries: dbQueries,
	}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Readiness (healthz) endpoint
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK\n"))
	})

	// Metrics endpoint
	mux.HandleFunc("GET /admin/metrics", apiCfg.HandleMetrics)

	// Reset endpoint
	mux.HandleFunc("POST /admin/reset", apiCfg.HandleReset)

	// Chirp Validation (validate_chirp) endpoint
	mux.HandleFunc("POST /api/validate_chirp", api.HandleValidChirp)

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
