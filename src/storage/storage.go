package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	primaryUrl := os.Getenv("TURSO_URL")
	if primaryUrl == "" {
		return nil, fmt.Errorf("TURSO_URL environment variable not set")
	}
	authToken := os.Getenv("TURSO_AUTH_TOKEN")
	if authToken == "" {
		return nil, fmt.Errorf("TURSO_AUTH_TOKEN environment variable not set")
	}

	url := fmt.Sprintf("libsql://%s?authToken=%s", primaryUrl, authToken)

	log.Printf("use turso: %s", primaryUrl)

	db, err := sql.Open("libsql", url)
	if err != nil {
		return nil, err
	}

	// TODO: https://foxcpp.dev/articles/the-right-way-to-use-go-sqlite3
	// db.SetMaxOpenConns(1)

	if err = migrate(db); err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}
