package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/papacatzzi-server/domain"
	"github.com/papacatzzi-server/email"
	"github.com/papacatzzi-server/postgres"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

const (
	EmailVerified = "EMAIL_VERIFIED"

	accessTokenExpiry  = time.Hour * 24 * 7
	refreshTokenExpiry = time.Hour * 24 * 30
)

type AuthService struct {
	repository postgres.UserRepository
	redis      *redis.Client
	mailer     email.Mailer
}

func NewAuthService(repo postgres.UserRepository, redis *redis.Client, mailer email.Mailer) AuthService {
	return AuthService{repository: repo, redis: redis, mailer: mailer}
}

func (svc *AuthService) Login(email string, password string) (accessToken string, refreshToken string, err error) {
	user, err := svc.repository.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrInvalidCredentials
			return
		}
		err = fmt.Errorf("failed to fetch username from db: %v", err)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		err = domain.ErrInvalidCredentials
		return
	}

	accessToken, err = createToken(user, accessTokenExpiry)
	if err != nil {
		err = fmt.Errorf("failed to create access token: %v", err)
		return
	}

	refreshToken, err = createToken(user, refreshTokenExpiry)
	if err != nil {
		err = fmt.Errorf("failed to create refresh token: %v", err)
		return
	}

	return
}

func (svc *AuthService) BeginSignUp(email string) (err error) {
	// check if there is an active account with this email
	user, err := svc.repository.GetUserByEmail(email)
	if user.IsActive {
		err = domain.ErrUserAccountActive
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("failed to fetch email from db: %v", err)
		return
	}

	// cache verification code into redis
	code := generateVerificationCode()

	err = svc.redis.Set(context.Background(), email, code, time.Minute*5).Err()
	if err != nil {
		err = fmt.Errorf("failed to cache verification code: %v", err)
		return
	}

	go func() {
		if err := svc.mailer.SendVerificationCode(email, code); err != nil {
			// TODO: Add logging
			return
		}
	}()

	return
}

func (svc *AuthService) VerifySignUp(email string, code string) (err error) {
	// check cache if verification code is correct
	cached, err := svc.redis.Get(context.Background(), email).Result()
	if err != nil {
		err = fmt.Errorf("failed to get verification code: %v", err)
		return
	}

	if cached != code {
		err = domain.ErrIncorrectCode
		return
	}

	// if successful, set status to verified
	err = svc.redis.Set(context.Background(), email, EmailVerified, time.Minute*5).Err()
	if err != nil {
		err = fmt.Errorf("failed to cache email verification status: %v", err)
		return
	}

	return
}

func (svc *AuthService) FinishSignUp(email string, username string, password string) (err error) {
	// check cache if email was verified
	status, err := svc.redis.Get(context.Background(), email).Result()
	if err != nil || status != EmailVerified {
		err = fmt.Errorf("failed to get verification status: %v", err)
		return
	}

	_, err = svc.repository.GetUserByName(username)
	if err == nil {
		err = domain.ErrUsernameExists
		return
	}

	if !errors.Is(err, sql.ErrNoRows) {
		err = fmt.Errorf("failed to fetch username from db: %v", err)
		return
	}

	// hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return
	}

	// save user's email
	newUser := domain.User{
		Email:     email,
		Username:  username,
		Password:  string(hashed),
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	err = svc.repository.InsertUser(newUser)
	if err != nil {
		err = fmt.Errorf("failed to insert user: %v", err)
		return
	}

	// clean up cache after new user is saved
	_, err = svc.redis.Del(context.Background(), email).Result()
	if err != nil {
		// TODO: Add logging
	}

	return
}

func generateVerificationCode() (code string) {
	digits := "0123456789"
	for i := 0; i < 6; i++ {
		code += string(digits[rand.Intn(len(digits))])
	}

	return code
}

type claims struct {
	jwt.RegisteredClaims
	Email string
}

func createToken(user domain.User, expiry time.Duration) (string, error) {
	claims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
		Email: user.Email,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
