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

// TODO: get coordinates DTO
func (s *Server) listSightings(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	minLng, err := strconv.ParseFloat(queryParams.Get("minLng"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing minLng", http.StatusBadRequest)
		return
	}

	minLat, err := strconv.ParseFloat(queryParams.Get("minLat"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing minLat", http.StatusBadRequest)
		return
	}

	maxLng, err := strconv.ParseFloat(queryParams.Get("maxLng"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing maxLng", http.StatusBadRequest)
		return
	}

	maxLat, err := strconv.ParseFloat(queryParams.Get("maxLat"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing maxLat", http.StatusBadRequest)
		return
	}

	sightings, err := s.store.GetSightingsByCoordinates(minLng, minLat, maxLng, maxLat)
	if err != nil {
		http.Error(w, "Error getting sightings.", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(sightings)
}

// TODO: get details DTO
func (s *Server) getSighting(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	sighting, err := s.store.GetSightingByID(id)
	if err != nil {
		http.Error(w, "Error getting sighting by ID.", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(sighting)
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
		http.Error(w, "Error creating sighting.", http.StatusUnprocessableEntity)
		return
	}

	if err := csr.Validate(); err != nil {
		log.Print("failed to validate create sighting request: ", err)
		http.Error(w, "Error creating sighting", http.StatusUnprocessableEntity)
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
		http.Error(w, "Error creating sighting.", http.StatusUnprocessableEntity)
		return
	}

	w.WriteHeader(http.StatusOK)
}
