package file

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/kirill555525/URL/internal/logger"
	"github.com/kirill555525/URL/internal/models"
	"github.com/kirill555525/URL/internal/store"
	"github.com/kirill555525/URL/internal/utils"
	"go.uber.org/zap"
	"io"
	"os"
	"strconv"
	"sync"
)

type Store struct {
	Path   string
	Memory store.Store
	Mutex  sync.Mutex
	Count  int
}

type URLStruct struct {
	ID          string `json:"uuid"`
	ShortID     string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewStore(path string, storage store.Store) (*Store, error) {
	s := &Store{
		Path:   path,
		Memory: storage,
		Mutex:  sync.Mutex{},
		Count:  0,
	}

	err := s.readURLFile()
	if err != nil {
		logger.Log.Error("error reading file", zap.Error(err))
		return nil, err
	}

	return s, nil
}

func (s *Store) Close() {
	s.Memory.Close()
}

func (s *Store) CheckConnectDB(ctx context.Context) error {
	return s.Memory.CheckConnectDB(ctx)
}

func (s *Store) readURLFile() error {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	file, err := os.OpenFile(s.Path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	for {
		obj := URLStruct{}
		err = decoder.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		_, err = s.Memory.SaveIDAndURL(nil, obj.ShortID, obj.OriginalURL)
		if err != nil {
			return err
		}
		s.Count++

	}

	return nil
}

func (s *Store) GetOriginalURL(ctx context.Context, shortID string) (originalURL string, err error) {
	return s.Memory.GetOriginalURL(ctx, shortID)
}

func (s *Store) GetShortID(ctx context.Context, originalURL string) (shortID string, err error) {
	return s.Memory.GetShortID(ctx, originalURL)
}

func (s *Store) SaveIDAndURL(ctx context.Context, shortID string, originalURL string) (string, error) {
	s.Mutex.Lock()
	defer s.Mutex.Unlock()
	res, err := s.Memory.SaveIDAndURL(ctx, shortID, originalURL)

	if err != nil {
		return res, err
	}

	file, err := os.OpenFile(s.Path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return "", err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	obj := URLStruct{
		ID:          strconv.Itoa(s.Count + 1),
		ShortID:     shortID,
		OriginalURL: originalURL,
	}
	err = encoder.Encode(&obj)
	if err != nil {
		return "", err
	}
	s.Count++
	return "", nil

}

func (s *Store) SaveIDAndURLJSONList(ctx context.Context, shortIDList []string, originalURLJSONList []models.RequestBatchURL) ([]models.ResponseBatchURL, error) {

	s.Mutex.Lock()
	defer s.Mutex.Unlock()

	file, err := os.OpenFile(s.Path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)

	response := make([]models.ResponseBatchURL, len(shortIDList))
	for i, shortID := range shortIDList {
		originalURL := originalURLJSONList[i].OriginalURL

		res, err := s.Memory.SaveIDAndURL(ctx, shortID, originalURL)
		var isWrite bool = true

		if errors.Is(err, store.ErrURLAlreadyExists) {
			shortID = res
			isWrite = false
		} else if err != nil {
			return nil, err
		}

		if isWrite {
			obj := URLStruct{
				ID:          strconv.Itoa(s.Count + 1),
				ShortID:     shortID,
				OriginalURL: originalURL,
			}
			err = encoder.Encode(&obj)
			if err != nil {
				return nil, err
			}
			s.Count++
		}

		shortURL := utils.CreateShortURL(shortID)
		response[i].CorrelationID = originalURLJSONList[i].CorrelationID
		response[i].ShortURL = shortURL

	}
	return response, nil

}
