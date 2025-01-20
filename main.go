package main

import (
	database "github.com/papacatzzi-server/db"
	"github.com/papacatzzi-server/email"
	"github.com/papacatzzi-server/http"
	"github.com/papacatzzi-server/log"
	"github.com/papacatzzi-server/postgres"
	"github.com/papacatzzi-server/service"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := log.NewLogger()

	mailer := email.NewMailer()

	db, err := database.NewDB()
	if err != nil {
		logger.Fatal().Err(err)
		return
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	sightingRepo := postgres.NewSightingRepository(db)
	userRepo := postgres.NewUserRepository(db)

	sightingService := service.NewSightingService(sightingRepo)
	authService := service.NewAuthService(userRepo, rdb, mailer)

	server := http.NewServer(logger, authService, sightingService)
	server.ListenAndServe()
}
