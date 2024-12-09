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

	r.HandleFunc("/posts", s.getPosts).Methods("GET")
	return
}

func (s *Server) ListenAndServe() {
	log.Fatal(s.server.ListenAndServe())
}
