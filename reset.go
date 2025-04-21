package main

import "net/http"

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(-1)
	w.Header().Set("Content-Type", "text/plain; charset=utf-9")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metrics reset to -1"))
}
