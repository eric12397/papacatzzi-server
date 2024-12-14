package db

import (
	"github.com/papacatzzi-server/models"
)

func (s Store) GetCoordinates(
	minLng float64,
	minLat float64,
	maxLng float64,
	maxLat float64,
) (coordinates []models.Coordinates, err error) {

	rows, err := s.db.Query(`
		SELECT id, latitude, longitude, created_at
		FROM posts
		WHERE ST_MakePoint(longitude, latitude) && ST_MakeEnvelope($1, $2, $3, $4, 4326)
	`, minLng, minLat, maxLng, maxLat)

	if err != nil {
		return
	}

	for rows.Next() {
		var coords models.Coordinates
		if err = rows.Scan(&coords.PostID, &coords.Latitude, &coords.Longitude, &coords.Timestamp); err != nil {
			return
		}

		coordinates = append(coordinates, coords)
	}

	return
}
