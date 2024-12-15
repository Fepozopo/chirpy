package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

func main() {
	os.Exit(main_helper())
}

func main_helper() int {
	filepathRoot := "."
	port := "8080"

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Custom FileServer to handle all routes dynamically
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(filepathRoot, r.URL.Path)
		file, err := os.Open(filePath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer file.Close()

		// Get file information
		fileInfo, err := file.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Dynamically set the Content-Length header based on the file size
		w.Header().Set("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

		// Serve the file
		http.ServeContent(w, r, filePath, fileInfo.ModTime(), file)
	}))

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
