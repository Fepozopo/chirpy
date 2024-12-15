package main

import (
	"net/http"
)

// middlewareMetricsInc wraps the given http.Handler and increments the fileserverHits
// counter on each request.
func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}