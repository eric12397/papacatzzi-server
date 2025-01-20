package postgres

import (
	"database/sql"

	"github.com/papacatzzi-server/domain"
)

type SightingRepository struct {
	db *sql.DB
}

func NewSightingRepository(db *sql.DB) SightingRepository {
	return SightingRepository{db: db}
}

func (r SightingRepository) GetSightingsByCoordinates(
	minLng float64,
	minLat float64,
	maxLng float64,
	maxLat float64,
) (sightings []domain.Sighting, err error) {

	rows, err := r.db.Query(`
		SELECT id, latitude, longitude, created_at
		FROM sightings
		WHERE ST_MakePoint(longitude, latitude) && ST_MakeEnvelope($1, $2, $3, $4, 4326)
	`, minLng, minLat, maxLng, maxLat)

	if err != nil {
		return
	}

	for rows.Next() {
		var s domain.Sighting
		rows.Scan(&s.ID, &s.Latitude, &s.Longitude, &s.Timestamp)
		if err != nil {
			return
		}

		sightings = append(sightings, s)
	}

	return
}

func (r SightingRepository) GetSightingByID(id string) (sighting domain.Sighting, err error) {

	err = r.db.QueryRow(`
		SELECT id, user_id, animal_type, photo_url, description, created_at
		FROM sightings
		WHERE id = $1
	`, id).Scan(&sighting.ID, &sighting.Reporter, &sighting.Animal, &sighting.PhotoURL, &sighting.Description, &sighting.Timestamp)

	return
}

func (r SightingRepository) InsertSighting(sighting domain.Sighting) (err error) {

	_, err = r.db.Exec(`
		INSERT INTO sightings 
		(user_id, animal_type, photo_url, description, latitude, longitude, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, sighting.Reporter, sighting.Animal, sighting.PhotoURL, sighting.Description, sighting.Latitude, sighting.Longitude, sighting.Timestamp)

	return
}
