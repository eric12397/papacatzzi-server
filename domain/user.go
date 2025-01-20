package domain

import "time"

type User struct {
	ID        int
	Username  string
	Email     string
	Password  string
	CreatedAt time.Time
	IsActive  bool
}

// TODO: define possible interfaces for service/repo here
