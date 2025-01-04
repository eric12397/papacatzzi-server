package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/papacatzzi-server/models"
	"golang.org/x/crypto/bcrypt"
)

type signUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req signUpRequest) Validate() (err error) {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Username, validation.Required),
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required, validation.Length(10, 20)),
	)
}

func (s *Server) signUp(w http.ResponseWriter, r *http.Request) {
	var req signUpRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Print("failed to parse request json: ", err)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
		return
	}

	if err := req.Validate(); err != nil {
		log.Print("failed to validate sign up request: ", err)
		http.Error(w, "Error signing up new user", http.StatusUnprocessableEntity)
		return
	}

	_, err := s.store.GetUserByName(req.Username)
	if err == nil {
		log.Print("username exists: ", req.Username)
		http.Error(w, "Error signing up new user", http.StatusUnprocessableEntity)
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		log.Print("error fetching username from db: ", req.Username)
		http.Error(w, "Error signing up new user", http.StatusUnprocessableEntity)
		return
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return
	}

	newUser := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashed),
		CreatedAt: time.Now(),
	}

	err = s.store.InsertUser(newUser)
	if err != nil {
		log.Print("failed to insert user: ", err)
		http.Error(w, "Error creating user.", http.StatusUnprocessableEntity)
		return
	}

	// TODO: return jwt
	w.WriteHeader(http.StatusOK)
}
