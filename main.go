package main

import (
	"github.com/papacatzzi-server/api"
	"github.com/papacatzzi-server/db"
)

func main() {
	db := db.NewStore()

	server, err := api.NewServer(db)
	if err != nil {
		return
	}

	server.ListenAndServe()
}
