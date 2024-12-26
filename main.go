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

	// Open a connection to the database and environment variables
	godotenv.Load()
	tokenSecret := os.Getenv("TOKEN_SECRET")
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Printf("Failed to open a connection to the database: %v\n", err)
	}
	defer db.Close()

	// Create a new Queries instance and initialize the ApiConfig struct
	dbQueries := database.New(db)
	apiCfg := &api.ApiConfig{
		DbQueries:   dbQueries,
		TokenSecret: tokenSecret,
	}

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("GET /api/healthz", apiCfg.HandleHealthz)
	mux.HandleFunc("GET /admin/metrics", apiCfg.HandleMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.HandleReset)
	mux.HandleFunc("POST /api/chirps", apiCfg.HandleCreateChirp)
	mux.HandleFunc("POST /api/users", apiCfg.HandleCreateUser)
	mux.HandleFunc("GET /api/chirps", apiCfg.HandleGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.HandleGetChirp)
	mux.HandleFunc("POST /api/login", apiCfg.HandleLoginUser)
	mux.HandleFunc("POST /api/refresh", apiCfg.HandleRefresh)
	mux.HandleFunc("POST /api/revoke", apiCfg.HandleRevoke)
	mux.HandleFunc("PUT /api/users", apiCfg.HandleUpdateUser)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.HandleDeleteChirp)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.HandleStripeEvent)

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
