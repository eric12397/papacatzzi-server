package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getCoordinates(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	northEastLat, err := strconv.ParseFloat(queryParams.Get("northEastLat"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing northEastLat", http.StatusBadRequest)
		return
	}

	northEastLng, err := strconv.ParseFloat(queryParams.Get("northEastLng"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing northEastLng", http.StatusBadRequest)
		return
	}

	southWestLat, err := strconv.ParseFloat(queryParams.Get("southWestLat"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing southWestLat", http.StatusBadRequest)
		return
	}

	southWestLng, err := strconv.ParseFloat(queryParams.Get("southWestLng"), 64)
	if err != nil {
		http.Error(w, "Invalid or missing southWestLng", http.StatusBadRequest)
		return
	}

	posts, err := s.store.GetCoordinates(northEastLat, northEastLng, southWestLat, southWestLng)
	if err != nil {
		http.Error(w, "Error getting posts.", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(posts)
}
