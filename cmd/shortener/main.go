package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/kirill555525/URL/cmd/config"
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

		var shortID string

		if url, ok := urlMap[longURL]; ok {
			shortID = url
		} else {

			shortID, err = generateShortURL()

			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			urlMap[longURL] = shortID
			idMap[shortID] = longURL
		}

		shortURL := fmt.Sprintf("%s%s", cfg.BaseURL, shortID)

		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(shortURL))

	}
}

func URLRouter() chi.Router {

	cfg := config.GetConfig()

	router := chi.NewRouter()
	router.Get("/{shortID}", shortenURLHandlerGet)
	router.Post("/", shortenURLHandlerPost(cfg))

	return router
}

func main() {
	cfg := config.Init()

	if err := http.ListenAndServe(cfg.Addr, URLRouter()); err != nil {
		panic(err)
	}
}
