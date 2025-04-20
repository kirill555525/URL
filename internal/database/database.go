package database

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/models"
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

// проверить generateShortURL()

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

func GetORCreateShortURLList(ctx context.Context, req []models.RequestBatchURL) ([]models.ResponseBatchURL, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	checkURLQuery := `SELECT short_url FROM url WHERE original_url = $1`
	insertURLQuery := `INSERT INTO url (uuid, short_url, original_url) VALUES ($1, $2, $3) RETURNING short_url`

	result := make([]models.ResponseBatchURL, 0, 1000)
	cfg := config.GetConfig()

	defer tx.Rollback()

	for _, obj := range req {
		var shortID string
		err := tx.QueryRowContext(ctx, checkURLQuery, obj.OriginalURL).Scan(&shortID)

		if errors.Is(err, sql.ErrNoRows) {
			shortID, err = generateShortURL()
			if err != nil {
				return nil, err
			}
			id := uuid.New()

			err = tx.QueryRowContext(ctx, insertURLQuery, id, shortID, obj.OriginalURL).Scan(&shortID) // можно переписать на exec
			if err != nil {
				return nil, fmt.Errorf("failed to insert new URL: %w", err)
			}

		} else if err != nil {
			return nil, fmt.Errorf("failed to query URL: %w", err)
		}

		result = append(result, models.ResponseBatchURL{CorrelationID: obj.CorrelationID, ShortURL: fmt.Sprintf("%s/%s", cfg.BaseURL, shortID)})
	}

	err = tx.Commit()
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return result, nil

}

func generateShortURL() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	res := base64.URLEncoding.EncodeToString(b)

	return res, nil
}
