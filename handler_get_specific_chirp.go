package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerGetSpecificChirp(w http.ResponseWriter, r *http.Request) {
	chirpIDString := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDString)
	if err != nil {
		fmt.Println("Error parsing chirpID.")
		respondWithError(w, http.StatusBadRequest, (`{"error": "Invalid chirp id."}`))
		return
	}

	chirp, err := cfg.DB.GetSpecificChirp(r.Context(), chirpID)
	if err != nil {
		fmt.Println("Error fetching specific Chirp with GetSpecificChirp. The chirpID is: ", chirpID)
		respondWithError(w, http.StatusNotFound, (`{"error": "Error fetching specific chirp."}`))
		return
	}

	response := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		UserID:    chirp.UserID,
		Body:      chirp.Body,
	}

	chirpJson, err := json.Marshal(response)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, (`{"error": "Error marshalling chirp."}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(chirpJson))

}
