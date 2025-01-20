package domain

import "time"

type Sighting struct {
	ID          int
	Animal      string
	Description string
	PhotoURL    string
	Reporter    string

	Latitude  float64
	Longitude float64
	Timestamp time.Time
}

// TODO: define possible interfaces for service/repo here
