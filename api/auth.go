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

const EmailVerified = "EMAIL_VERIFIED"

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

	// check if there is an active account with this email
	user, err := s.store.GetUserByEmail(req.Email)
	if user.IsActive {
		log.Print("user with this email exists: ", req.Email)
		http.Error(w, "User with this email exists", http.StatusUnprocessableEntity)
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

	go func() {
		if err := s.mailer.SendVerificationCode(req.Email, code); err != nil {
			log.Print("failed to send verification code via email: ", err)
			return
		}
	}()

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
		validation.Field(&req.Code, validation.Required, validation.Length(6, 6), is.Digit),
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

	// check cache if verification code is correct
	cached, err := s.redis.Get(context.Background(), req.Email).Result()
	if err != nil {
		log.Print("failed to get verification code: ", err)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
		return
	}

	if cached != req.Code {
		log.Print("incorrect code: ", req.Code)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
		return
	}

	// if successful, set status to verified
	err = s.redis.Set(context.Background(), req.Email, EmailVerified, time.Minute*5).Err()
	if err != nil {
		log.Print("failed to cache verification code: ", err)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
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

	// check cache if email was verified
	status, err := s.redis.Get(context.Background(), req.Email).Result()
	if err != nil || status != EmailVerified {
		log.Print("failed to get verification status: ", err)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
		return
	}

	_, err = s.store.GetUserByName(req.Username)
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

	// save user's email
	newUser := models.User{
		Email:     req.Email,
		Username:  req.Username,
		Password:  string(hashed),
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	err = s.store.InsertUser(newUser)
	if err != nil {
		log.Print("failed to insert user: ", err)
		http.Error(w, "Error creating user.", http.StatusUnprocessableEntity)
		return
	}

	// clean up cache after new user is saved
	_, err = s.redis.Del(context.Background(), req.Email).Result()
	if err != nil {
		log.Print("failed to delete verification status: ", err)
		http.Error(w, "Error signing up new user.", http.StatusUnprocessableEntity)
		return
	}

	// TODO: return jwt
	log.Print("user signed up successfully: ", newUser.Username)
	w.WriteHeader(http.StatusOK)
}
