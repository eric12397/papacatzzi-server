package http

import (
	"encoding/json"
	"errors"
	"net/http"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/papacatzzi-server/domain"
)

const EmailVerified = "EMAIL_VERIFIED"

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (req loginRequest) Validate() (err error) {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Email, validation.Required, is.Email),
		validation.Field(&req.Password, validation.Required),
	)
}

type loginResponse struct {
	AccessToken string `json:"access"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Error().Err(err).Msg("failed to parse request")
		s.errorResponse(w, http.StatusBadRequest, "Error parsing request")
		return
	}

	if err := req.Validate(); err != nil {
		s.errorResponse(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	accessToken, refreshToken, err := s.authService.Login(req.Email, req.Password)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		switch {
		case errors.Is(err, domain.ErrInvalidCredentials):
			s.errorResponse(w, http.StatusUnauthorized, "Invalid username or password")
		default:
			s.errorResponse(w, http.StatusInternalServerError, "Failed to process log in")
		}
		return
	}

	res := loginResponse{
		AccessToken: accessToken,
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,                    // Use Secure in production (requires HTTPS)
		SameSite: http.SameSiteStrictMode, // Adjust SameSite policy as needed
		//MaxAge: ,
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

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

	err := s.authService.BeginSignUp(req.Email)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		switch {
		case errors.Is(err, domain.ErrUserAccountActive):
			s.errorResponse(w, http.StatusBadRequest, err.Error())
		default:
			s.errorResponse(w, http.StatusInternalServerError, "failed to begin sign up")
		}
		return
	}

	s.logger.Debug().Msg("email sent successfully")
	w.WriteHeader(http.StatusOK)
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

	err := s.authService.VerifySignUp(req.Email, req.Code)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		switch {
		case errors.Is(err, domain.ErrIncorrectCode):
			s.errorResponse(w, http.StatusBadRequest, err.Error())
		default:
			s.errorResponse(w, http.StatusInternalServerError, "failed to verify sign up")
		}
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

	err := s.authService.FinishSignUp(req.Email, req.Username, req.Password)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		switch {
		case errors.Is(err, domain.ErrUsernameExists):
			s.errorResponse(w, http.StatusBadRequest, err.Error())
		default:
			s.errorResponse(w, http.StatusInternalServerError, "failed to finish sign up")
		}
		return
	}

	// TODO: return jwt
	//s.logger.Debug().Msgf("user signed up successfully: %v", newUser.Username)
	w.WriteHeader(http.StatusOK)
}

type refreshTokenResponse struct {
	AccessToken string `json:"access"`
}

func (s *Server) refreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := r.Cookie("refresh_token")
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid or missing cookie")
		return
	}

	accessToken, err := s.authService.RefreshToken(refreshToken.Value)
	if err != nil {
		s.logger.Error().Msg(err.Error())
		s.errorResponse(w, http.StatusInternalServerError, "Error refreshing token")
		return
	}

	res := refreshTokenResponse{
		AccessToken: accessToken,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
