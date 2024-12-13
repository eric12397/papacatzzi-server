package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getCoordinates(w http.ResponseWriter, r *http.Request) {
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

	coords, err := s.store.GetCoordinates(minLng, minLat, maxLng, maxLat)
	if err != nil {
		http.Error(w, "Error getting coords.", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(coords)
}
