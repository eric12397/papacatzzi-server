package main

import (
	"github.com/papacatzzi-server/api"
	"github.com/papacatzzi-server/db"
)

func main() {
	store, err := db.NewStore()
	if err != nil {
		return
	}

	server, err := api.NewServer(store)
	if err != nil {
		return
	}

	server.ListenAndServe()
}
