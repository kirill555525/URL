package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/compress"
	"github.com/kirill555525/URL/internal/database"
	"github.com/kirill555525/URL/internal/logger"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// PingGetHandler проверяет соединение с базой данных
func PingGetHandler(w http.ResponseWriter, r *http.Request) {
	if err := database.CheckConnectDB(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

const shortURLLength = 6 // 8 символов в base64

func ReadURLFile() error {
	cfg := config.GetConfig()

	if cfg.FileStoragePath == "" {
		return nil
	}

	file, err := os.OpenFile(cfg.FileStoragePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)

	obj := URLStruct{}

	for {
		err = decoder.Decode(&obj)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		idMap[obj.ShortURL] = obj.OriginalURL
		urlMap[obj.OriginalURL] = obj.ShortURL

	}

	return nil
}

func WriteURLFile(shortID string) error {
	cfg := config.GetConfig()

	if cfg.FileStoragePath == "" {
		return nil
	}

	file, err := os.OpenFile(cfg.FileStoragePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	uuid := len(idMap)
	res := URLStruct{
		ID:          strconv.Itoa(uuid),
		ShortURL:    shortID,
		OriginalURL: idMap[shortID],
	}
	err = encoder.Encode(&res)
	return err
}

type URLStruct struct {
	ID          string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

var (
	urlMap = make(map[string]string)
	idMap  = make(map[string]string)
	mutex  sync.Mutex
)

func generateShortURL() (string, error) {
	for {
		b := make([]byte, shortURLLength)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		res := base64.URLEncoding.EncodeToString(b)

		_, exists := idMap[res]

		if !exists {
			return res, nil
		}

	}

}

func URLRouter() chi.Router {

	router := chi.NewRouter()

	router.Use(logger.ResponseLogger)
	router.Use(logger.RequestLogger)
	router.Use(compress.GzipMiddleware)

	router.Get("/{shortID}", shortenURLHandlerGet)
	router.Post("/", shortenURLHandlerPost)
	router.Post("/api/shorten", APIShortenHandlerPost)
	router.Get("/ping", PingGetHandler)
	router.Post("/api/shorten/batch", BatchShortenHandlerPost)
	return router
}

func main() {
	cfg := config.Init()
	err := logger.Initialize(cfg.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	if cfg.DatabaseDSN != "" {
		db, err := database.ConnectDB()
		if err != nil {
			panic(err)
		}

		if err := database.CheckConnectDB(context.Background()); err != nil {
			panic(err)
		}

		if err := database.CreateTableURL(context.Background()); err != nil {
			panic(err)
		}

		defer db.Close()

	} else {
		err := ReadURLFile()
		if err != nil {
			panic(err)
		}
	}

	if err := http.ListenAndServe(cfg.Addr, URLRouter()); err != nil {
		panic(err)
	}
}
