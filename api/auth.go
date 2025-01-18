package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
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
		s.logger.Error().Err(err).Msg("failed to parse request")
		s.errorResponse(w, http.StatusBadRequest, "Error parsing request")
		return
	}

	if err := req.Validate(); err != nil {
		s.errorResponse(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// check if there is an active account with this email
	user, err := s.store.GetUserByEmail(req.Email)
	if user.IsActive {
		s.errorResponse(w, http.StatusBadRequest, "User with this email exists")
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error().Err(err).Msg("failed to fetch email from db")
		s.errorResponse(w, http.StatusInternalServerError, "Error signing up new user")
		return
	}

	// cache verification code into redis
	code := generateVerificationCode()

	err = s.redis.Set(context.Background(), req.Email, code, time.Minute*5).Err()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to cache verification code")
		s.errorResponse(w, http.StatusInternalServerError, "Error creating verification code")
		return
	}

	go func() {
		if err := s.mailer.SendVerificationCode(req.Email, code); err != nil {
			s.logger.Error().Err(err).Msg("failed to send verification code via email")
			return
		}
	}()

	s.logger.Debug().Msg("email sent successfully")
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
		s.logger.Error().Err(err).Msg("failed to parse request")
		s.errorResponse(w, http.StatusBadRequest, "Error parsing request")
		return
	}

	if err := req.Validate(); err != nil {
		s.errorResponse(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// check cache if verification code is correct
	cached, err := s.redis.Get(context.Background(), req.Email).Result()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to get verification code")
		s.errorResponse(
			w,
			http.StatusInternalServerError,
			"Error verifying email. Please start over to generate a new verification code.",
		)
		return
	}

	if cached != req.Code {
		s.errorResponse(w, http.StatusUnprocessableEntity, "Incorrect code")
		return
	}

	// if successful, set status to verified
	err = s.redis.Set(context.Background(), req.Email, EmailVerified, time.Minute*5).Err()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to cache email verification status")
		s.errorResponse(w, http.StatusInternalServerError, "Error verifying email")
		return
	}

	s.logger.Debug().Msg("email verified successfully")
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
		s.logger.Error().Err(err).Msg("failed to parse request")
		s.errorResponse(w, http.StatusBadRequest, "Error parsing request")
		return
	}

	if err := req.Validate(); err != nil {
		s.errorResponse(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// check cache if email was verified
	status, err := s.redis.Get(context.Background(), req.Email).Result()
	if err != nil || status != EmailVerified {
		s.logger.Error().Err(err).Msg("failed to get verification status")
		s.errorResponse(
			w,
			http.StatusInternalServerError,
			"Error creating new account. Please start over sign up flow.",
		)
		return
	}

	_, err = s.store.GetUserByName(req.Username)
	if err == nil {
		s.errorResponse(w, http.StatusBadRequest, "Username already exists")
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error().Err(err).Msg("failed to fetch username from db")
		s.errorResponse(w, http.StatusInternalServerError, "Error verifying username")
		return
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		s.errorResponse(w, http.StatusInternalServerError, err.Error())
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
		s.logger.Error().Err(err).Msg("failed to insert user")
		s.errorResponse(w, http.StatusInternalServerError, "Error saving new account")
		return
	}

	// clean up cache after new user is saved
	_, err = s.redis.Del(context.Background(), req.Email).Result()
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to delete verification status")
	}

	// TODO: return jwt
	s.logger.Debug().Msgf("user signed up successfully: %v", newUser.Username)
	w.WriteHeader(http.StatusOK)
}
