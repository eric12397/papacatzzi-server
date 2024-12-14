package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}

func NewStore() (store Store, err error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")

	dataSourceName := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
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
