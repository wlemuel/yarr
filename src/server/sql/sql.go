package sql

import (
	"database/sql"
	"io"
	"log"
)

func Execute(r io.Reader, db *sql.DB) error {
	content, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	log.Print(string(content))
	_, err = db.Exec(string(content))
	if err != nil {
		return err
	}
	return nil
}
