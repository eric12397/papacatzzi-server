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
	"github.com/google/uuid"
	"github.com/papacatzzi-server/domain"
	smtp "github.com/papacatzzi-server/email"
	"github.com/papacatzzi-server/postgres"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

const (
	SignUpVerificationKey = "SIGN_UP_VERIFICATION"
	VerificationCompleted = "VERIFICATION_COMPLETED"

	BlacklistKey     = "BLACKLIST"
	TokenBlacklisted = "TOKEN_BLACKLISTED"

	AccessTokenExpiration        = time.Minute * 15
	RefreshTokenExpiration       = time.Hour * 24 * 7
	PasswordResetTokenExpiration = time.Hour
)

type AuthService struct {
	repository postgres.UserRepository
	redis      *redis.Client
	mailer     smtp.Mailer
}

func NewAuthService(repo postgres.UserRepository, redis *redis.Client, mailer smtp.Mailer) AuthService {
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

	accessToken, err = createToken(user, AccessTokenExpiration)
	if err != nil {
		err = fmt.Errorf("failed to create access token: %v", err)
		return
	}

	refreshToken, err = createToken(user, RefreshTokenExpiration)
	if err != nil {
		err = fmt.Errorf("failed to create refresh token: %v", err)
		return
	}

	// TODO: store refresh token in redis, revoke/delete when user logs out

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
	code := generateCode(6)
	key := appendToKey(SignUpVerificationKey, email)

	err = svc.redis.Set(context.Background(), key, code, time.Minute*5).Err()
	if err != nil {
		err = fmt.Errorf("failed to cache verification code: %v", err)
		return
	}

	go func() {
		data := map[string]string{}

		content := smtp.EmailContent{
			Subject:   "Sign up your new account",
			Recipient: email,
			Body:      data,
		}

		if err := svc.mailer.Send("email/templates/signup.html", content); err != nil {
			// TODO: Add logging
			return
		}
	}()

	return
}

func (svc *AuthService) VerifySignUp(email string, code string) (err error) {
	// check cache if verification code is correct
	key := appendToKey(SignUpVerificationKey, email)

	cached, err := svc.redis.Get(context.Background(), key).Result()
	if err != nil {
		err = fmt.Errorf("failed to get verification code: %v", err)
		return
	}

	if cached != code {
		err = domain.ErrIncorrectCode
		return
	}

	// if successful, set status to verified
	err = svc.redis.Set(context.Background(), key, VerificationCompleted, time.Minute*5).Err()
	if err != nil {
		err = fmt.Errorf("failed to cache email verification status: %v", err)
		return
	}

	return
}

func (svc *AuthService) FinishSignUp(email string, username string, password string) (err error) {
	// check cache if email was verified
	key := appendToKey(SignUpVerificationKey, email)

	status, err := svc.redis.Get(context.Background(), key).Result()
	if err != nil || status != VerificationCompleted {
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
	_, err = svc.redis.Del(context.Background(), key).Result()
	if err != nil {
		// TODO: Add logging
	}

	return
}

func (svc *AuthService) ForgotPassword(email string) (err error) {
	// check if there is an active account with this email
	user, err := svc.repository.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrUserAccountNotFound
			return
		}

		err = fmt.Errorf("failed to fetch user from db: %v", err)
		return
	}

	token, err := createToken(user, PasswordResetTokenExpiration)
	if err != nil {
		err = fmt.Errorf("failed to create password reset token: %v", err)
		return
	}

	go func() {
		data := map[string]string{
			"username": user.Username,
			"link":     fmt.Sprintf("http://localhost:5173/reset-password?token=%v", token),
		}

		content := smtp.EmailContent{
			Subject:   "Password reset",
			Recipient: email,
			Body:      data,
		}

		if err := svc.mailer.Send("email/templates/password-reset.html", content); err != nil {
			// TODO: Add logging
			return
		}
	}()

	return
}

func (svc *AuthService) ResetPassword(token string, password string) (err error) {
	// get email from token
	claims, err := svc.VerifyToken(token)
	if err != nil {
		return
	}

	user, err := svc.repository.GetUserByEmail(claims.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrUserAccountNotFound
			return
		}

		err = fmt.Errorf("failed to fetch user from db: %v", err)
		return
	}

	// validate that the new password does not match old one
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err == nil {
		err = domain.ErrSamePassword
		return
	}

	// hash new password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return
	}

	err = svc.repository.UpdatePassword(hashed, user.Email)
	if err != nil {
		err = fmt.Errorf("failed to update password: %v", err)
		return
	}

	return
}

func (svc *AuthService) VerifyToken(tokenString string) (c *claims, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		err = fmt.Errorf("error parsing token, %v", err)
		return
	}

	claims, ok := token.Claims.(*claims)
	if !ok || !token.Valid {
		err = fmt.Errorf("invalid token")
		return
	}

	c = claims
	return
}

func (svc *AuthService) RefreshToken(refreshToken string) (accessToken string, err error) {
	claims, err := svc.VerifyToken(refreshToken)
	if err != nil {
		return
	}

	user, err := svc.repository.GetUserByEmail(claims.Email)
	if err != nil {
		err = fmt.Errorf("failed to fetch username from db: %v", err)
		return
	}

	accessToken, err = createToken(user, AccessTokenExpiration)
	if err != nil {
		err = fmt.Errorf("failed to create access token: %v", err)
		return
	}

	return
}

func (svc *AuthService) CompleteOAuth(oAuthID string, email string) (accessToken string, refreshToken string, err error) {

	user, err := svc.repository.GetUserByEmail(email)
	if errors.Is(err, sql.ErrNoRows) {

		// fallback to finding by oauth id
		user, err = svc.repository.GetUserByOAuthID(oAuthID)
		if errors.Is(err, sql.ErrNoRows) {

			// create if not found
			user = domain.User{
				OAuthID:   oAuthID,
				Email:     email,
				CreatedAt: time.Now(),
				IsActive:  true,
			}

			// auto generate username for oauth users, can update later
			user.Username = "AnonymousUser" + generateCode(12)

			err = svc.repository.InsertUser(user)
			if err != nil {
				err = fmt.Errorf("failed to insert user: %v", err)
				return
			}

		} else if err != nil {
			err = fmt.Errorf("failed to fetch user by oauth from db: %v", err)
			return
		}

	} else if err != nil {
		err = fmt.Errorf("failed to fetch user by email from db: %v", err)
		return
	}

	// link missing oauth id in case user originally signed up for account via normal flow
	if user.OAuthID == "" {
		if err = svc.repository.UpdateOAuthID(oAuthID, user.Email); err != nil {
			err = fmt.Errorf("failed to update oauth id from db: %v", err)
			return
		}
	}

	accessToken, err = createToken(user, AccessTokenExpiration)
	if err != nil {
		err = fmt.Errorf("failed to create access token: %v", err)
		return
	}

	refreshToken, err = createToken(user, RefreshTokenExpiration)
	if err != nil {
		err = fmt.Errorf("failed to create refresh token: %v", err)
		return
	}

	return
}

func generateCode(length int) (code string) {
	digits := "0123456789"
	for i := 0; i < length; i++ {
		code += string(digits[rand.Intn(len(digits))])
	}

	return code
}

type claims struct {
	jwt.RegisteredClaims
	Email  string    `json:"email"`
	UserID uuid.UUID `json:"id"`
}

func createToken(user domain.User, expiration time.Duration) (string, error) {
	claims := claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
		},
		Email:  user.Email,
		UserID: user.ID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func appendToKey(key string, id string) string {
	return key + "_" + id
}
