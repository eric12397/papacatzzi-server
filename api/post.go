package api

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) getPosts(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	north, err := strconv.ParseFloat(queryParams.Get("north"), 64)
	if err != nil {
		http.Error(w, "Error parsing coordinates.", http.StatusBadRequest)
		return
	}

	south, err := strconv.ParseFloat(queryParams.Get("south"), 64)
	if err != nil {
		http.Error(w, "Error parsing coordinates.", http.StatusBadRequest)
		return
	}

	east, err := strconv.ParseFloat(queryParams.Get("east"), 64)
	if err != nil {
		http.Error(w, "Error parsing coordinates.", http.StatusBadRequest)
		return
	}

	west, err := strconv.ParseFloat(queryParams.Get("west"), 64)
	if err != nil {
		http.Error(w, "Error parsing coordinates.", http.StatusBadRequest)
		return
	}

	posts, err := s.store.GetPosts(north, south, east, west)
	if err != nil {
		http.Error(w, "Error getting posts.", http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(posts)
}
