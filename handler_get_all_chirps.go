package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (cfg *apiConfig) handlerGetAllChirps(w http.ResponseWriter, r *http.Request) {
	response, err := cfg.DB.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 400, (`{"error": "Something went wrong when retrieving all chirps."}`))
		return
	}

	allChirps := []Chirp{}
	for _, dbChirp := range response {
		allChirps = append(allChirps, Chirp{
			ID:        dbChirp.ID,
			CreatedAt: dbChirp.CreatedAt,
			UpdatedAt: dbChirp.UpdatedAt,
			UserID:    dbChirp.UserID,
			Body:      dbChirp.Body,
		})
	}

	data, err := json.Marshal(allChirps)
	if err != nil {
		fmt.Println("Error marshalling all chirps!")
		return
	}
	w.WriteHeader(200)
	w.Write([]byte(data))
}
