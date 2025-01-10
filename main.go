package main

import (
	"github.com/papacatzzi-server/api"
	"github.com/papacatzzi-server/db"
	"github.com/papacatzzi-server/email"
	"github.com/redis/go-redis/v9"
)

func main() {
	store, err := db.NewStore()
	if err != nil {
		return
	}

	mailer := email.NewMailer()

	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	server, err := api.NewServer(store, mailer, rdb)
	if err != nil {
		return
	}

	server.ListenAndServe()
}
