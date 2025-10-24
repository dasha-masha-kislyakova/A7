package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func MustConnect(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("db ping: %v", err)
	}
	return db
}
