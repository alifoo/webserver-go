package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alifoo/webserver-go/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
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

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.PLATFORM != "dev" {
		respondWithError(w, 403, (`{"error": "Platform is not dev."}`))
		return
	}

	err := cfg.DB.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, 400, (`{"error": "Something went wrong with deleting all users"}`))
		return
	}

	cfg.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Metrics reset to 0"))

}

func (cfg *apiConfig) CreateUsersEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	type parameters struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, (`{"error": "Something went wrong"}`))
		return
	}

	dbUser, err := cfg.DB.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 400, (`{"error": "Something went wrong with user creation"}`))
		return
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	data, err := json.Marshal(user)
	if err != nil {
		respondWithError(w, 500, `{"error": "Error marshalling JSON"}`)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write(data)
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
	PLATFORM       string
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
	platform := os.Getenv("PLATFORM")
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
		DB:       dbQueries,
		PLATFORM: platform,
	}
	handler := http.StripPrefix("/app/", http.FileServer(http.Dir("./")))

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(handler))
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsReader)
	mux.HandleFunc("POST /api/validate_chirp", ValidateChirpEndpoint)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUsersEndpoint)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	fmt.Println("Server up and running!")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
