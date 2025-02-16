package main

import (
	"os"

	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
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

	goth.UseProviders(
		google.New(os.Getenv("GOOGLE_CLIENT_ID"), os.Getenv("GOOGLE_CLIENT_SECRET"), os.Getenv("GOOGLE_CLIENT_CALLBACK_URL")),
	)

	sightingRepo := postgres.NewSightingRepository(db)
	userRepo := postgres.NewUserRepository(db)

	sightingService := service.NewSightingService(sightingRepo)
	authService := service.NewAuthService(userRepo, rdb, mailer)

	server := http.NewServer(logger, authService, sightingService)
	server.ListenAndServe()
}
