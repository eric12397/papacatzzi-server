package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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
