package store

import (
	"context"
	"errors"
	"github.com/kirill555525/URL/internal/models"
)

var ErrNotFound = errors.New("record not found")
var ErrURLAlreadyExists = errors.New("url already exists")
var ErrShortIDAlreadyExists = errors.New("shortID already exists")

type Store interface {
	GetOriginalURL(ctx context.Context, shortID string) (originalURL string, err error)
	GetShortID(ctx context.Context, originalURL string) (shortID string, err error)
	SaveIDAndURL(ctx context.Context, shortID string, originalURL string) (string, error)
	SaveIDAndURLJSONList(ctx context.Context, shortIDList []string, originalURLJSONList []models.RequestBatchURL) ([]models.ResponseBatchURL, error)
	Close()
	CheckConnectDB(ctx context.Context) error
}
