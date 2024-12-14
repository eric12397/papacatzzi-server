package models

import "time"

type Post struct {
	Animal   string
	PhotoURL string
	Author   string
}

type Coordinates struct {
	PostID    int
	Latitude  float64
	Longitude float64
	Timestamp time.Time
}
