package db

import (
	"time"

	"github.com/papacatzzi-server/models"
)

func (s Store) GetCoordinates(
	northEastLat float64,
	northEastLng float64,
	southWestLat float64,
	southWestLng float64,
) (coords []models.Coordinates, err error) {

	coords = append(coords,
		models.Coordinates{
			Latitude:  northEastLat,
			Longitude: northEastLng,
			Timestamp: time.Now(),
		},
		models.Coordinates{
			Latitude:  southWestLat,
			Longitude: southWestLng,
			Timestamp: time.Now(),
		},
	)

	return
}
