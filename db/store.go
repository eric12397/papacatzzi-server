package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "papacatzzi"
	password = "papacatzzi"
	dbname   = "papacatzzi"
)

type Store struct {
	db *sql.DB
}

func NewStore() (store Store, err error) {
	dataSourceName := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return
	}

	// TODO: input db connection
	store = Store{db}

	return
}
