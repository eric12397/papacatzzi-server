package domain

import "errors"

var (
	ErrUserAccountActive = errors.New("user's account is already active")
	ErrUsernameExists    = errors.New("username already exists")
	ErrIncorrectCode     = errors.New("incorrect verification code")
)
