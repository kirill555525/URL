package store_test

import (
	"context"
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/logger"
	"github.com/kirill555525/URL/internal/store"
	"github.com/kirill555525/URL/internal/store/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMemory(t *testing.T) {
	cfg := config.Init()
	err := logger.Initialize(cfg.FlagLogLevel)
	require.NoError(t, err)
	//storage, err := postgres.NewStore(cfg.DatabaseDSN)
	//require.NoError(t, err)
	storage := memory.NewStore()
	//mem := memory.NewStore()
	//storage := file.NewStore(cfg.FileStoragePath, mem)

	defer storage.Close()

	tests := []struct {
		name        string
		method      string
		shortID     string
		originalURL string
		result      string
		wantErr     error
	}{
		{
			name:        "GetOriginalURL NOT EXIST",
			method:      "GetOriginalURL",
			shortID:     "sdca9801",
			originalURL: "",
			result:      "",
			wantErr:     store.ErrNotFound,
		},
		{
			name:        "GetShortID NOT EXIST",
			method:      "GetShortID",
			shortID:     "",
			originalURL: "https://www.google.ru/",
			result:      "",
			wantErr:     store.ErrNotFound,
		},
		{
			name:        "SaveIDAndURL #1",
			method:      "SaveIDAndURL",
			shortID:     "hello",
			originalURL: "https://www.google.ru/",
			result:      "",
			wantErr:     nil,
		},
		{
			name:        "SaveIDAndURL #2",
			method:      "SaveIDAndURL",
			shortID:     "hhhhhhhh",
			originalURL: "https://ya.ru/",
			result:      "",
			wantErr:     nil,
		},
		{
			name:        "SaveIDAndURL URLAlreadyExists",
			method:      "SaveIDAndURL",
			shortID:     "hhhhhhha",
			originalURL: "https://ya.ru/",
			result:      "hhhhhhhh",
			wantErr:     store.ErrURLAlreadyExists,
		},
		{
			name:        "SaveIDAndURL URLAlreadyExists and ShortIDAlreadyExists",
			method:      "SaveIDAndURL",
			shortID:     "hhhhhhhs",
			originalURL: "https://ya.ru/",
			result:      "hhhhhhhh",
			wantErr:     store.ErrURLAlreadyExists,
		},
		{
			name:        "SaveIDAndURL ShortIDAlreadyExists",
			method:      "SaveIDAndURL",
			shortID:     "hello",
			originalURL: "https://go.dev/",
			result:      "",
			wantErr:     store.ErrShortIDAlreadyExists,
		},
		{
			name:        "GetOriginalURL",
			method:      "GetOriginalURL",
			shortID:     "hello",
			originalURL: "",
			result:      "https://www.google.ru/",
			wantErr:     nil,
		},
		{
			name:        "GetShortID",
			method:      "GetShortID",
			shortID:     "",
			originalURL: "https://ya.ru/",
			result:      "hhhhhhhh",
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var res string
			var err error

			switch tt.method {
			case "GetOriginalURL":
				res, err = storage.GetOriginalURL(context.Background(), tt.shortID)
			case "GetShortID":
				res, err = storage.GetShortID(context.Background(), tt.originalURL)
			case "SaveIDAndURL":
				res, err = storage.SaveIDAndURL(context.Background(), tt.shortID, tt.originalURL)
			}
			assert.Equal(t, tt.result, res)
			assert.Equal(t, tt.wantErr, err)
		})
	}

}
