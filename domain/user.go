package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	OAuthID   string
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	IsActive  bool
}

// TODO: define possible interfaces for service/repo here
