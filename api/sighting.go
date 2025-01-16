package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gorilla/mux"
	"github.com/papacatzzi-server/models"
)

type coordinates struct {
	ID        int       `json:"id"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Timestamp time.Time `json:"timestamp"`
}

func (s *Server) listSightings(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	minLng, err := strconv.ParseFloat(queryParams.Get("minLng"), 64)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid or missing minLat")
		return
	}

	minLat, err := strconv.ParseFloat(queryParams.Get("minLat"), 64)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid or missing minLat")
		return
	}

	maxLng, err := strconv.ParseFloat(queryParams.Get("maxLng"), 64)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid or missing maxLng")
		return
	}

	maxLat, err := strconv.ParseFloat(queryParams.Get("maxLat"), 64)
	if err != nil {
		s.errorResponse(w, http.StatusBadRequest, "Invalid or missing maxLat")
		return
	}

	sightings, err := s.store.GetSightingsByCoordinates(minLng, minLat, maxLng, maxLat)
	if err != nil {
		s.errorResponse(w, http.StatusNotFound, "Sightings not found at specified coordinates")
		return
	}

	coords := make([]coordinates, 0)
	for _, s := range sightings {
		coords = append(coords, coordinates{
			ID:        s.ID,
			Latitude:  s.Latitude,
			Longitude: s.Longitude,
			Timestamp: s.Timestamp,
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(coords)
}

type sightingDetailsResponse struct {
	Animal      string    `json:"animal"`
	Description string    `json:"description"`
	PhotoURL    string    `json:"photoURL"`
	Reporter    string    `json:"reporter"`
	Timestamp   time.Time `json:"timestamp"`
}

func (s *Server) getSighting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sighting, err := s.store.GetSightingByID(id)
	if err != nil {
		log.Print("failed to get sighting by id: ", err)
		s.errorResponse(w, http.StatusNotFound, "Sighting not found by ID")
		return
	}

	res := sightingDetailsResponse{
		Animal:      sighting.Animal,
		Description: sighting.Description,
		PhotoURL:    sighting.PhotoURL,
		Reporter:    sighting.Reporter,
		Timestamp:   sighting.Timestamp,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

type createSightingRequest struct {
	Animal      string    `json:"animal"`
	Description string    `json:"description"`
	PhotoURL    string    `json:"photoURL"`
	Reporter    string    `json:"reporter"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	Timestamp   time.Time `json:"timestamp"`
}

func (csr createSightingRequest) Validate() (err error) {
	// Leaving out reporter for now...
	// Need to implement users/accounts
	return validation.ValidateStruct(&csr,
		validation.Field(&csr.Animal, validation.Required),
		validation.Field(&csr.Description, validation.Required),
		validation.Field(&csr.PhotoURL, validation.Required),
		validation.Field(&csr.Latitude, validation.Required),
		validation.Field(&csr.Longitude, validation.Required),
		validation.Field(&csr.Timestamp, validation.Required),
	)
}

func (s *Server) createSighting(w http.ResponseWriter, r *http.Request) {
	var csr createSightingRequest

	if err := json.NewDecoder(r.Body).Decode(&csr); err != nil {
		log.Print("failed to parse request json: ", err)
		s.errorResponse(w, http.StatusBadRequest, "Error parsing request")
		return
	}

	if err := csr.Validate(); err != nil {
		log.Print("failed to validate create sighting request: ", err)
		s.errorResponse(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	newSighting := models.Sighting{
		Animal:      csr.Animal,
		Description: csr.Description,
		PhotoURL:    csr.PhotoURL,
		Reporter:    csr.Reporter,
		Latitude:    csr.Latitude,
		Longitude:   csr.Longitude,
		Timestamp:   csr.Timestamp,
	}

	err := s.store.InsertSighting(newSighting)
	if err != nil {
		log.Print("failed to insert sighting: ", err)
		s.errorResponse(w, http.StatusInternalServerError, "Error creating sighting")
		return
	}

	w.WriteHeader(http.StatusOK)
}
