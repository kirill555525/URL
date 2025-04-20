package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/database"
	"github.com/kirill555525/URL/internal/models"
	"io"
	"net/http"
	"strings"
)

func shortenURLHandlerGet(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")

	if database.IsDBUsed() {
		url, err := database.ReadURL(r.Context(), shortID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
		return

	}

	mutex.Lock()
	defer mutex.Unlock()

	if url, ok := idMap[shortID]; ok {

		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusTemporaryRedirect)
	} else {
		http.Error(w, "URL Not Found", http.StatusNotFound)
	}

}

func getORCreateShortURLNODB(longURL string) (string, error) {
	cfg := config.GetConfig()
	mutex.Lock()
	defer mutex.Unlock()

	var shortID string

	if url, ok := urlMap[longURL]; ok {
		shortID = url
	} else {
		var err error
		shortID, err = generateShortURL()

		if err != nil {
			return "", err
		}

		urlMap[longURL] = shortID
		idMap[shortID] = longURL

		err = WriteURLFile(shortID)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%s/%s", cfg.BaseURL, shortID), nil
}

func getORCreateShortURL(ctx context.Context, longURL string) (string, error) {
	cfg := config.GetConfig()

	if database.IsDBUsed() {
		shortID, err := database.ReadShortID(ctx, longURL)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				shortID, err = generateShortURL()
				if err != nil {
					return "", err
				}

				err = database.WriteURL(ctx, shortID, longURL)
				if err != nil {
					return "", err
				}

			} else {
				return "", err
			}
		}
		return fmt.Sprintf("%s/%s", cfg.BaseURL, shortID), nil
	}

	return getORCreateShortURLNODB(longURL)
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

	shortURL, err := getORCreateShortURL(r.Context(), longURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

	shortURL, err := getORCreateShortURL(r.Context(), longURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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

func BatchShortenHandlerPost(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	t, err := decoder.Token()
	if err != nil || t != json.Delim('[') {
		http.Error(w, "expected JSON array", http.StatusBadRequest)
		return
	}

	items := make([]models.RequestBatchURL, 0, 1000)
	result := make([]models.ResponseBatchURL, 0, 1000)

	for decoder.More() {
		var item models.RequestBatchURL

		err := decoder.Decode(&item)
		if err != nil {
			http.Error(w, "invalid JSON object in array", http.StatusBadRequest)
			return
		}

		if database.IsDBUsed() {
			items = append(items, item)

			if len(items) == 1000 {
				resp, err := database.GetORCreateShortURLList(r.Context(), items)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				items = items[:0]
				result = append(result, resp...)
			}
		} else {
			longURL, err := getORCreateShortURLNODB(item.OriginalURL)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			result = append(result, models.ResponseBatchURL{
				CorrelationID: item.CorrelationID,
				ShortURL:      longURL,
			})
		}

	}

	if database.IsDBUsed() {
		resp, err := database.GetORCreateShortURLList(r.Context(), items)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		result = append(result, resp...)
	}

	_, err = decoder.Token()
	if err != nil {
		http.Error(w, "invalid JSON array end", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encode := json.NewEncoder(w)
	err = encode.Encode(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

}
