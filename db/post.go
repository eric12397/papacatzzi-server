package db

import (
	"time"

	"github.com/papacatzzi-server/models"
)

func (s Store) GetCoordinates(
	minLat float64,
	minLng float64,
	maxLat float64,
	maxLng float64,
) (coords []models.Coordinates, err error) {

	coords = append(coords,
		models.Coordinates{
			Latitude:  minLat,
			Longitude: minLng,
			Timestamp: time.Now(),
		},
		models.Coordinates{
			Latitude:  maxLat,
			Longitude: maxLng,
			Timestamp: time.Now(),
		},
	)

	return
}
