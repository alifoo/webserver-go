package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/alifoo/webserver-go/internal/database"
	"github.com/google/uuid"
	"time"
)

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	type parameters struct {
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, (`{"error": "Something went wrong while decoding params."}`))
		fmt.Println(err)
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

	chirpParams := database.CreateChirpParams{
		Body:   joinedWords,
		UserID: params.UserId,
	}

	newChirp, err := cfg.DB.CreateChirp(r.Context(), chirpParams)
	if err != nil {
		respondWithError(w, 400, (`{"error": "Something went wrong with chirp creation"}`))
		return
	}

	response := Chirp{
		ID:        newChirp.ID,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
		Body:      newChirp.Body,
		UserID:    newChirp.UserID,
	}

	chirpJson, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshalling chirp!")
		return
	}

	w.WriteHeader(201)
	w.Write([]byte(chirpJson))

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
