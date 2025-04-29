package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kirill555525/URL/internal/models"
	"github.com/kirill555525/URL/internal/store"
	"github.com/kirill555525/URL/internal/utils"
	"time"
)

type Store struct {
	db *pgxpool.Pool
}

func NewStore(source string) (*Store, error) {
	db, err := pgxpool.New(context.Background(), source)
	if err != nil {
		return nil, err
	}

	s := &Store{db: db}
	err = s.CheckConnectDB(context.Background())
	if err != nil {
		return nil, err
	}

	err = s.createTableURLS(context.Background())
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) createTableURLS(ctx context.Context) error {

	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_id VARCHAR(8) UNIQUE NOT NULL,
		original_url TEXT UNIQUE NOT NULL
	);`
	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	_, err := s.db.Exec(ctxChild, query)
	if err != nil {
		return err
	}

	return nil

}

func (s *Store) CheckConnectDB(ctx context.Context) error {

	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()

	if err := s.db.Ping(ctxChild); err != nil {
		return err
	}

	return nil

}

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) GetOriginalURL(ctx context.Context, shortID string) (originalURL string, err error) {

	query := `
		SELECT original_url FROM urls WHERE short_id = $1;
	`
	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	row := s.db.QueryRow(ctxChild, query, shortID)

	err = row.Scan(&originalURL)

	if errors.Is(err, sql.ErrNoRows) {
		return "", store.ErrNotFound
	} else if err != nil {
		return "", err
	}

	return originalURL, nil

}

func (s *Store) GetShortID(ctx context.Context, originalURL string) (shortID string, err error) {

	query := `
		SELECT short_id FROM urls WHERE original_url = $1;
	`
	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	row := s.db.QueryRow(ctxChild, query, originalURL)

	err = row.Scan(&shortID)

	if errors.Is(err, sql.ErrNoRows) {
		return "", store.ErrNotFound
	} else if err != nil {
		return "", err
	}
	return shortID, nil

}

// если ключ и значение совпадают - у меня разное поведение. у memory и этого. сам сценарий маловероятен. Приведи ку единому интерфейсу

func (s *Store) SaveIDAndURL(ctx context.Context, shortID string, originalURL string) (string, error) {

	query := `
	INSERT INTO urls (short_id, original_url)
	VALUES ($1, $2)
	ON CONFLICT (original_url)
	DO UPDATE SET 
    	short_id = urls.short_id 
	RETURNING short_id;`

	ctxChild, cancel := context.WithTimeout(ctx, time.Second*3)
	defer cancel()
	row := s.db.QueryRow(ctxChild, query, shortID, originalURL)

	var id string

	err := row.Scan(&id)

	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return "", store.ErrShortIDAlreadyExists
			}
		}
		fmt.Printf("%T\n", err)

		return "", err
	}

	if id != shortID {
		return id, store.ErrURLAlreadyExists
	}

	return "", nil

}

func (s *Store) SaveIDAndURLJSONList(ctx context.Context, shortIDList []string, originalURLJSONList []models.RequestBatchURL) ([]models.ResponseBatchURL, error) {
	response := make([]models.ResponseBatchURL, len(shortIDList))
	for i, shortID := range shortIDList {
		originalURL := originalURLJSONList[i].OriginalURL

		res, err := s.SaveIDAndURL(ctx, shortID, originalURL)

		if errors.Is(err, store.ErrURLAlreadyExists) {
			shortID = res
		} else if err != nil {
			return nil, err
		}

		shortURL := utils.CreateShortURL(shortID)
		response[i].CorrelationID = originalURLJSONList[i].CorrelationID
		response[i].ShortURL = shortURL

	}
	return response, nil
}
