package db

import (
	"github.com/papacatzzi-server/models"
)

func (s Store) GetSightingsByCoordinates(
	minLng float64,
	minLat float64,
	maxLng float64,
	maxLat float64,
) (sightings []models.Sighting, err error) {

	rows, err := s.db.Query(`
		SELECT id, latitude, longitude, created_at
		FROM sightings
		WHERE ST_MakePoint(longitude, latitude) && ST_MakeEnvelope($1, $2, $3, $4, 4326)
	`, minLng, minLat, maxLng, maxLat)

	if err != nil {
		return
	}

	for rows.Next() {
		var s models.Sighting
		rows.Scan(&s.ID, &s.Latitude, &s.Longitude, &s.Timestamp)
		if err != nil {
			return
		}

		sightings = append(sightings, s)
	}

	return
}

func (s Store) GetSightingByID(id string) (sighting models.Sighting, err error) {

	err = s.db.QueryRow(`
		SELECT id, user_id, animal_type, photo_url, description, created_at
		FROM sightings
		WHERE id = $1
	`, id).Scan(&sighting.ID, &sighting.Reporter, &sighting.Animal, &sighting.PhotoURL, &sighting.Description, &sighting.Timestamp)

	return
}

func (s Store) InsertSighting(sighting models.Sighting) (err error) {

	_, err = s.db.Exec(`
		INSERT INTO sightings 
		(user_id, animal_type, photo_url, description, latitude, longitude, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, sighting.Reporter, sighting.Animal, sighting.PhotoURL, sighting.Description, sighting.Latitude, sighting.Longitude, sighting.Timestamp)

	return
}
