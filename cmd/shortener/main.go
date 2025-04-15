package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/compress"
	"github.com/kirill555525/URL/internal/logger"
	"github.com/kirill555525/URL/internal/models"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

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

func shortenURLHandlerGet(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")

	mutex.Lock()
	defer mutex.Unlock()

	if url, ok := idMap[shortID]; ok {

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "URL Not Found", http.StatusNotFound)
	}

}

func shortenURLHandlerPost(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if len(body) == 0 {
		http.Error(w, "Empty body", http.StatusBadRequest)
		return
	}

	longURL := strings.TrimSpace(string(body))

	mutex.Lock()
	defer mutex.Unlock()
	cfg := config.GetConfig()

	var shortID string

	if url, ok := urlMap[longURL]; ok {
		shortID = url
	} else {

		shortID, err = generateShortURL()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urlMap[longURL] = shortID
		idMap[shortID] = longURL

		err = WriteURLFile(shortID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, shortID)

	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))

}

func APIShortenHandlerPost(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var req models.Request

	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	longURL := req.URL

	mutex.Lock()
	defer mutex.Unlock()
	cfg := config.GetConfig()

	var shortID string

	if url, ok := urlMap[longURL]; ok {
		shortID = url
	} else {

		shortID, err = generateShortURL()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		urlMap[longURL] = shortID
		idMap[shortID] = longURL

		err = WriteURLFile(shortID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, shortID)

	resp := models.Response{
		ShortURL: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	err = encoder.Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
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

	return router
}

func main() {
	cfg := config.Init()
	err := ReadURLFile()
	if err != nil {
		panic(err)
	}
	err = logger.Initialize(cfg.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	if err := http.ListenAndServe(cfg.Addr, URLRouter()); err != nil {
		panic(err)
	}
}
