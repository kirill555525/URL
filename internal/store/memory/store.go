package memory

import (
	"context"
	"errors"
	"github.com/kirill555525/URL/internal/models"
	"github.com/kirill555525/URL/internal/store"
	"github.com/kirill555525/URL/internal/utils"
	"sync"
)

type Store struct {
	URLByID map[string]string
	IDByURL map[string]string
	Mutex   sync.Mutex
}

func NewStore() *Store {
	return &Store{
		URLByID: make(map[string]string),
		IDByURL: make(map[string]string),
		Mutex:   sync.Mutex{},
	}
}

func (s *Store) Close() {
	
}

func (s *Store) CheckConnectDB(ctx context.Context) error {
	return errors.New("the database is not in use")
}

func (s *Store) GetOriginalURL(ctx context.Context, shortID string) (originalURL string, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	url, ok := s.URLByID[shortID]
	if !ok {
		return "", store.ErrNotFound
	}

	return url, nil
}

func (s *Store) GetShortID(ctx context.Context, originalURL string) (shortID string, err error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	id, ok := s.IDByURL[originalURL]
	if !ok {
		return "", store.ErrNotFound
	}
	return id, nil
}

func (s *Store) SaveIDAndURL(ctx context.Context, shortID string, originalURL string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	id, ok := s.IDByURL[originalURL]
	if ok {
		return id, store.ErrURLAlreadyExists
	}
	_, ok = s.URLByID[shortID]
	if ok {
		return "", store.ErrShortIDAlreadyExists
	}

	s.IDByURL[originalURL] = shortID
	s.URLByID[shortID] = originalURL
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
