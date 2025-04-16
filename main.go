package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alifoo/webserver-go/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

func ReadinessEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte("OK"))

}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

func IsBadWord(word string) bool {
	switch word {
	case
		"kerfuffle",
		"sharbert",
		"fornax":
		return true
	}
	return false
}

func ReplaceBadWord(word string) string {
	//var censuredWord bytes.Buffer

	//	for i := 0; i < len(word); i++ {
	//		censuredWord.WriteString("*")
	//	}

	// return censuredWord.String()
	return "****"
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
		respondWithError(w, 400, (`{"error": "Something went wrong"}`))
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, (`{"error": "Chirp is too long"}`))
		return
	}

	currentWord := ""
	words := strings.Split(params.Body, " ")
	fmt.Println("Words before: ", words)

	for i := range words {
		currentWord = strings.ToLower(words[i])
		if IsBadWord(currentWord) {
			words[i] = ReplaceBadWord(currentWord)
		}
	}
	fmt.Println("Words after: ", words)
	joinedWords := strings.Join(words, " ")

	w.WriteHeader(200)
	jsonString := fmt.Sprintf(`{"cleaned_body": "%s"}`, joinedWords)
	w.Write([]byte(jsonString))

}

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
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
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	mux.HandleFunc("GET /api/healthz", ReadinessEndpoint)
	apiCfg := apiConfig{
		DB: dbQueries,
	}
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir("./")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsReader)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetrics)
	mux.HandleFunc("POST /api/validate_chirp", ValidateChirpEndpoint)

	fmt.Println("Server up and running!")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
