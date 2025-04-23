package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/alifoo/webserver-go/internal/auth"
	"github.com/alifoo/webserver-go/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Password  string    `json:"password`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

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

	hashedPassword, err := auth.HashPassword(params.Password)
	createUserParams := database.CreateUserParams{
		Email:    params.Email,
		Password: hashedPassword,
	}

	dbUser, err := cfg.DB.CreateUser(r.Context(), createUserParams)
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
