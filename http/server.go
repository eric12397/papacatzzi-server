package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/papacatzzi-server/log"
	"github.com/papacatzzi-server/service"
)

type Server struct {
	server          *http.Server
	logger          log.Logger
	authService     service.AuthService
	sightingService service.SightingService
}

func NewServer(
	logger log.Logger,
	authService service.AuthService,
	sightingService service.SightingService,
) (s *Server) {

	s = &Server{
		server:          &http.Server{Addr: ":8080"},
		logger:          logger,
		authService:     authService,
		sightingService: sightingService,
	}

	s.server.Handler = s.setupRouter()
	return
}

func (s *Server) setupRouter() (r *mux.Router) {
	r = mux.NewRouter()

	r.HandleFunc("/login", s.login).Methods("POST")

	r.HandleFunc("/signup/begin", s.beginSignUp).Methods("POST")
	r.HandleFunc("/signup/verify", s.verifySignUp).Methods("POST")
	r.HandleFunc("/signup/finish", s.finishSignUp).Methods("POST")

	r.HandleFunc("/sightings", s.listSightings).Methods("GET")
	r.Handle("/sightings", s.auth(http.HandlerFunc(s.createSighting))).Methods("POST", "OPTIONS")
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

func (s *Server) auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			s.errorResponse(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			s.errorResponse(w, http.StatusUnauthorized, "Invalid authorization header")
			return
		}

		token := parts[1]
		err := s.authService.VerifyToken(token)
		if err != nil {
			s.logger.Error().Msg(err.Error())
			s.errorResponse(w, http.StatusUnauthorized, "Error verifying token")
			return
		}

		next.ServeHTTP(w, r)
	})
}
