package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getCoordinates(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	minLat, err := strconv.ParseFloat(queryParams.Get("minLat"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing minLat", http.StatusBadRequest)
		return
	}

	minLng, err := strconv.ParseFloat(queryParams.Get("minLng"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing minLng", http.StatusBadRequest)
		return
	}

	maxLat, err := strconv.ParseFloat(queryParams.Get("maxLat"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing maxLat", http.StatusBadRequest)
		return
	}

	maxLng, err := strconv.ParseFloat(queryParams.Get("maxLng"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing maxLng", http.StatusBadRequest)
		return
	}

	posts, err := s.store.GetCoordinates(minLat, minLng, maxLat, maxLng)
	if err != nil {
		http.Error(w, "Error getting posts.", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(posts)
}
