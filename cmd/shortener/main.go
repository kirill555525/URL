package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/logger"
	"github.com/kirill555525/URL/internal/models"
	"io"
	"net/http"
	"strings"
	"sync"
)

const shortURLLength = 6 // 8 символов в base64

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

func shortenURLHandlerPost(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		}

		shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, shortID)

		w.Header().Set("Content-Type", "text/plain")

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(shortURL))

	}
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

	longURL := req.Url

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
	}

	shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, shortID)

	resp := models.Response{
		ShortUrl: shortURL,
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusCreated)

	encoder := json.NewEncoder(w)
	err = encoder.Encode(resp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	return

}

func URLRouter() chi.Router {

	cfg := config.GetConfig()

	router := chi.NewRouter()
	router.Get("/{shortID}", logger.ResponseLogger(logger.RequestLogger(shortenURLHandlerGet)))
	router.Post("/", logger.ResponseLogger(logger.RequestLogger(shortenURLHandlerPost(cfg))))
	router.Post("/api/shorten", logger.ResponseLogger(logger.RequestLogger(APIShortenHandlerPost)))

	return router
}

func main() {
	cfg := config.Init()
	err := logger.Initialize(cfg.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	if err := http.ListenAndServe(cfg.Addr, URLRouter()); err != nil {
		panic(err)
	}
}
