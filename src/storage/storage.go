package storage

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

type Storage struct {
	db *sql.DB
}

func New(path string) (*Storage, error) {
	isDebug := false
	var db *sql.DB
	var err error

	primaryUrl := os.Getenv("TURSO_URL")
	if primaryUrl == "" {
		isDebug = true
	}

	if isDebug {
		log.Println("local debug mode: true")
		db, err = sql.Open("libsql", fmt.Sprintf("file:%s", path))

		if err != nil {
			log.Printf("failed to open db %s: %s", path, err)
			return nil, err
		}

		db.SetMaxOpenConns(1)
	} else {
		authToken := os.Getenv("TURSO_AUTH_TOKEN")
		if authToken == "" {
			return nil, fmt.Errorf("TURSO_AUTH_TOKEN environment variable not set")
		}

		url := fmt.Sprintf("libsql://%s?authToken=%s", primaryUrl, authToken)

		log.Printf("use turso: %s", primaryUrl)

		db, err = sql.Open("libsql", url)
		if err != nil {
			return nil, err
		}

	}

	if err = migrate(db); err != nil {
		return nil, err
	}
	return &Storage{db: db}, nil
}
