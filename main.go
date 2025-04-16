package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"
)

func ReadinessEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))

}

func ValidateChirpEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	type parameters struct {
		Body string `json:"body"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "Something went wrong"}`))
		return
	}

	if len(params.Body) > 140 {
		w.WriteHeader(400)
		w.Write([]byte(`{"error": "Chirp is too long"}`))
		return
	}

	w.WriteHeader(200)
	w.Write([]byte(`{"valid": true}`))

}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // handlerfunc turns any func with the right signature into a http handler
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsReader(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	hits := cfg.fileserverHits.Load()

	fmt.Fprintf(w, `
		<html>
			<body>
				<h1>Welcome, Chirpy Admin</h1>
				<p>Chirpy has been visited %d times!</p>
			</body>
		</html>
	`, hits)
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metrics reset to 0"))
}

func main() {
	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("GET /api/healthz", ReadinessEndpoint)
	apiCfg := apiConfig{}
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir("./")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsReader)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", ValidateChirpEndpoint)

	fmt.Println("Server up and running!")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
