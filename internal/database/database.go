package database

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"github.com/kirill555525/URL/cmd/config"
	"time"
)

var db *sql.DB

func ConnectDB() (*sql.DB, error) {
	cfg := config.GetConfig()

	dataBase, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	db = dataBase

	return db, nil
}

func GetDB() *sql.DB {
	return db
}

func IsDBUsed() bool {
	cfg := config.GetConfig()
	return cfg.DatabaseDSN != ""
}

func CheckConnectDB(ctx context.Context) error {

	if !IsDBUsed() {
		return errors.New("database DSN is empty")
	}

	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	if err := db.PingContext(ctxChild); err != nil {
		return err
	}

	return nil

}

func CreateTableURL(ctx context.Context) error {

	query := `
        CREATE TABLE IF NOT EXISTS url (
            uuid UUID PRIMARY KEY,
            short_url TEXT NOT NULL,
            original_url TEXT NOT NULL
        );
    `
	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	_, err := db.ExecContext(ctxChild, query)
	if err != nil {
		return err
	}

	return nil

}

func WriteURL(ctx context.Context, shortID string, url string) error {

	id := uuid.New()

	query := `
		INSERT INTO url (uuid, short_url, original_url)
		VALUES ($1, $2, $3);
	`

	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	_, err := db.ExecContext(ctxChild, query, id, shortID, url)
	if err != nil {
		return err
	}
	return nil
}

func ReadURL(ctx context.Context, shortID string) (string, error) {
	query := `
		SELECT original_url FROM url WHERE short_url = $1;
	`
	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	row := db.QueryRowContext(ctxChild, query, shortID)

	var url string

	err := row.Scan(&url)
	if err != nil {
		return "", err
	}
	return url, nil

}

func ReadShortID(ctx context.Context, url string) (string, error) {
	query := `
		SELECT short_url FROM url WHERE original_url = $1;
	`
	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	row := db.QueryRowContext(ctxChild, query, url)

	var shortID string

	err := row.Scan(&shortID)
	if err != nil {
		return "", err
	}
	return shortID, nil

}
