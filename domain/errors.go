package domain

import "errors"

var (
	ErrUserAccountActive   = errors.New("user's account is already active")
	ErrUserAccountNotFound = errors.New("user's account was not found")
	ErrUsernameExists      = errors.New("username already exists")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrSamePassword        = errors.New("new password cannot match your old password")

	ErrIncorrectCode = errors.New("incorrect verification code")
)
