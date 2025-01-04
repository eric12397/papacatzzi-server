package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/papacatzzi-server/db"
)

type Server struct {
	server *http.Server
	store  db.Store
}

func NewServer(store db.Store) (s *Server, err error) {
	s = &Server{
		server: &http.Server{Addr: ":8080"},
		store:  store,
	}

	s.server.Handler = s.setupRouter()
	return
}

func (s *Server) setupRouter() (r *mux.Router) {
	r = mux.NewRouter()

	r.HandleFunc("/signup", s.signUp).Methods("POST")

	r.HandleFunc("/sightings", s.listSightings).Methods("GET")
	r.HandleFunc("/sightings", s.createSighting).Methods("POST")
	r.HandleFunc("/sightings/{id}", s.getSighting).Methods("GET")
	r.Use(corsMiddleware)
	return
}

func (s *Server) ListenAndServe() {
	log.Fatal(s.server.ListenAndServe())
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Add("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		if r.Method == "OPTIONS" {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
