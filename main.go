package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	os.Exit(main_helper())
}

func main_helper() int {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Define a simple handler for the root path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Create a new http.Server struct
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start the server
	log.Printf("Starting server on :8080")
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Server failed to start: %v", err)
		return 1
	}

	// No errors, return 0
	return 0
}
