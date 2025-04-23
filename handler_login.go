package main

import (
	"encoding/json"
	"github.com/alifoo/webserver-go/internal/auth"
	"net/http"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}

	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, (`{"error": "Something went wrong"}`))
		return
	}

	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, (`{"error": "Incorrect email or password}`))
		return
	}

	err = auth.CheckPasswordHash(user.Password, params.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, (`{"error": "Incorrect email or password"}`))
		return
	}

	userWithoutPassw := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	data, err := json.Marshal(userWithoutPassw)
	if err != nil {
		respondWithError(w, 500, `{"error": "Error marshalling JSON"}`)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
