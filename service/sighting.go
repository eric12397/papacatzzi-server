package service

import (
	"github.com/papacatzzi-server/domain"
	"github.com/papacatzzi-server/postgres"
)

type SightingService struct {
	repository postgres.SightingRepository
}

func NewSightingService(repo postgres.SightingRepository) SightingService {
	return SightingService{repository: repo}
}

func (svc *SightingService) List(minLng, minLat, maxLng, maxLat float64) (sightings []domain.Sighting, err error) {
	return svc.repository.GetSightingsByCoordinates(minLng, minLat, maxLng, maxLat)
}

func (svc *SightingService) GetByID(id string) (sighting domain.Sighting, err error) {
	return svc.repository.GetSightingByID(id)
}

func (svc *SightingService) Create(sighting domain.Sighting) (err error) {
	return svc.repository.InsertSighting(sighting)
}
