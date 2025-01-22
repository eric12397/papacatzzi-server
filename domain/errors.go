package domain

import "errors"

var (
	ErrUserAccountActive  = errors.New("user's account is already active")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrIncorrectCode = errors.New("incorrect verification code")
)
