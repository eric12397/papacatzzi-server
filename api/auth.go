package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"math/rand"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/papacatzzi-server/models"
	"golang.org/x/crypto/bcrypt"
)

type beginSignUpRequest struct {
	Email string `json:"email"`
}

func (req beginSignUpRequest) Validate() (err error) {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Email, validation.Required, is.Email),
	)
}

func (s *Server) beginSignUp(w http.ResponseWriter, r *http.Request) {
	var req beginSignUpRequest

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

	// check if there is an existing account with this email
	_, err := s.store.GetUserByEmail(req.Email)
	if err == nil {
		log.Print("user with this email exists: ", req.Email)
		http.Error(w, "Error signing up new user", http.StatusUnprocessableEntity)
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		log.Print("error fetching email from db: ", req.Email)
		http.Error(w, "Error signing up new user", http.StatusUnprocessableEntity)
		return
	}

	// cache verification code into redis
	code := generateVerificationCode()

	err = s.redis.Set(context.Background(), req.Email, code, time.Minute*5).Err()
	if err != nil {
		log.Print("failed to cache verification code: ", err)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
		return
	}

	if err := s.mailer.SendVerificationCode(req.Email, code); err != nil {
		log.Print("failed to send verification code via email: ", err)
		http.Error(w, "Error sending email", http.StatusUnprocessableEntity)
		return
	}

	log.Print("email sent successfully")
	w.WriteHeader(http.StatusOK)
}

func generateVerificationCode() (code string) {
	digits := "0123456789"
	for i := 0; i < 6; i++ {
		code += string(digits[rand.Intn(len(digits))])
	}

	return code
}

type verifySignUpRequest struct {
	Code  string `json:"code"`
	Email string `json:"email"`
}

func (req verifySignUpRequest) Validate() (err error) {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Code, validation.Required, validation.Length(0, 6)),
	)
}

func (s *Server) verifySignUp(w http.ResponseWriter, r *http.Request) {
	var req verifySignUpRequest

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

	// check cache if verification code is correct and if expired
	cached, err := s.redis.Get(context.Background(), req.Email).Result()
	if err != nil {
		log.Print("failed to get verification code: ", err)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
		return
	}

	if cached != req.Code {
		log.Print("incorrect code: ", err)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
		return
	}

	// if successful, delete cache entry
	_, err = s.redis.Del(context.Background(), req.Email).Result()
	if err != nil {
		log.Print("failed to delete verification code from redis: ", err)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
		return
	}

	// save user's email
	newUser := models.User{
		Email:     req.Email,
		CreatedAt: time.Now(),
	}

	err = s.store.InsertUser(newUser)
	if err != nil {
		log.Print("failed to insert user: ", err)
		http.Error(w, "Error creating user.", http.StatusUnprocessableEntity)
		return
	}

	log.Print("email verified successfully")
	w.WriteHeader(http.StatusOK)
}

type finishSignUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req finishSignUpRequest) Validate() (err error) {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Username, validation.Required),
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required, validation.Length(10, 20)),
	)
}

func (s *Server) finishSignUp(w http.ResponseWriter, r *http.Request) {
	var req finishSignUpRequest

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

	// find user's email that was previously saved in the verify step
	user, err := s.store.GetUserByEmail(req.Email)
	if err != nil {
		log.Print("failed to find user's email: ", err)
		http.Error(w, "User did not verify email", http.StatusUnprocessableEntity)
		return
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return
	}

	user.Username = req.Username
	user.Password = string(hashed)
	user.IsActive = true

	err = s.store.ActivateUser(user)
	if err != nil {
		log.Print("failed to insert user: ", err)
		http.Error(w, "Error creating user.", http.StatusUnprocessableEntity)
		return
	}

	// TODO: return jwt
	log.Print("user signed up successfully: ", user.Username)
	w.WriteHeader(http.StatusOK)
}
