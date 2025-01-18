package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/papacatzzi-server/db"
	"github.com/papacatzzi-server/email"
	"github.com/papacatzzi-server/log"
	"github.com/redis/go-redis/v9"
)

type Server struct {
	server *http.Server
	store  db.Store
	mailer email.Mailer
	redis  *redis.Client
	logger log.Logger
}

func NewServer(store db.Store, mailer email.Mailer, redis *redis.Client, logger log.Logger) (s *Server, err error) {
	s = &Server{
		server: &http.Server{Addr: ":8080"},
		store:  store,
		mailer: mailer,
		redis:  redis,
		logger: logger,
	}

	s.server.Handler = s.setupRouter()
	return
}

func (s *Server) setupRouter() (r *mux.Router) {
	r = mux.NewRouter()

	r.HandleFunc("/signup/begin", s.beginSignUp).Methods("POST")
	r.HandleFunc("/signup/verify", s.verifySignUp).Methods("POST")
	r.HandleFunc("/signup/finish", s.finishSignUp).Methods("POST")

	r.HandleFunc("/sightings", s.listSightings).Methods("GET")
	r.HandleFunc("/sightings", s.createSighting).Methods("POST")
	r.HandleFunc("/sightings/{id}", s.getSighting).Methods("GET")
	r.Use(corsMiddleware)
	return
}

func (s *Server) ListenAndServe() {
	s.logger.Fatal().Err((s.server.ListenAndServe()))
}

func (s *Server) errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(message)
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
