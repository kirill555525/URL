package controllers

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/kirill555525/URL/internal/compress"
	"github.com/kirill555525/URL/internal/logger"
	"github.com/kirill555525/URL/internal/models"
	"github.com/kirill555525/URL/internal/store"
	"github.com/kirill555525/URL/internal/utils"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strings"
)

//type Logger interface {
//	Info(args ...interface{})
//}

type BaseController struct {
	storage store.Store // архитектура верхнего - нижнего слоя, вторая архитектура использовать app, третья архитетура использовать app как синглтон
	//logger Logger
}

func NewBaseController(storage store.Store) *BaseController {
	return &BaseController{storage: storage}
}

func (c *BaseController) Route() *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.ResponseLogger)
	r.Use(logger.RequestLogger)
	r.Use(compress.GzipMiddleware)
	r.Get("/{shortID}", c.getOriginalURL)
	r.Post("/", c.createShortURL)
	r.Post("/api/shorten", c.createShortURLJSON)
	r.Post("/api/shorten/batch", c.createShortURLJSONList)
	r.Get("/ping", c.PingDB)
	return r
}

func (c *BaseController) PingDB(w http.ResponseWriter, r *http.Request) {
	if err := c.storage.CheckConnectDB(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (c *BaseController) getOriginalURL(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "shortID")

	url, err := c.storage.GetOriginalURL(r.Context(), shortID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusTemporaryRedirect)

}

func (c *BaseController) createShortURL(w http.ResponseWriter, r *http.Request) {
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

	shortID, err := utils.GenerateShortID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortURL := utils.CreateShortURL(shortID)
	status := http.StatusCreated

	idIfExist, err := c.storage.SaveIDAndURL(r.Context(), shortID, longURL)
	if errors.Is(err, store.ErrURLAlreadyExists) {
		shortURL = utils.CreateShortURL(idIfExist)
		status = http.StatusConflict

	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(status)
	w.Write([]byte(shortURL))

}

func (c *BaseController) createShortURLJSON(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var req models.Request

	err := decoder.Decode(&req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	longURL := req.URL

	shortID, err := utils.GenerateShortID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	shortURL := utils.CreateShortURL(shortID)
	status := http.StatusCreated

	idIfExist, err := c.storage.SaveIDAndURL(r.Context(), shortID, longURL)
	if errors.Is(err, store.ErrURLAlreadyExists) {
		shortURL = utils.CreateShortURL(idIfExist)
		status = http.StatusConflict
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp := models.Response{
		ShortURL: shortURL,
	}

	res, err := json.Marshal(resp)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)
	w.Write(res)

}

func (c *BaseController) createShortURLJSONList(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	t, err := decoder.Token()
	if err != nil || t != json.Delim('[') {
		http.Error(w, "expected JSON array", http.StatusBadRequest)
		return
	}

	items := make([]models.RequestBatchURL, 0, 1000)
	shortIDList := make([]string, 0, 1000)
	result := make([]models.ResponseBatchURL, 0, 100)

	for decoder.More() {
		var item models.RequestBatchURL

		err := decoder.Decode(&item)
		if err != nil {
			http.Error(w, "invalid JSON object in array", http.StatusBadRequest)
			return
		}

		items = append(items, item)
		shortID, err := utils.GenerateShortID()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		shortIDList = append(shortIDList, shortID)

		if len(items) == 1000 {
			res, err := c.storage.SaveIDAndURLJSONList(r.Context(), shortIDList, items)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			items = items[:0]
			shortIDList = shortIDList[:0]
			result = append(result, res...)
		}

	}

	if len(items) > 0 {
		res, err := c.storage.SaveIDAndURLJSONList(r.Context(), shortIDList, items)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		result = append(result, res...)
	}

	if len(result) == 0 {
		http.Error(w, "empty list is not allowed", http.StatusBadRequest)
		return
	}

	t, err = decoder.Token()
	if err != nil || t != json.Delim(']') {
		http.Error(w, "invalid JSON array end", http.StatusBadRequest) // может стоит написать в тело, то что было обработано и добавлено
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	encode := json.NewEncoder(w) // возможно стоит перенести запись в конец каждой итерации len(1000)
	err = encode.Encode(result)
	if err != nil {
		logger.Log.Debug("", zap.Error(err))
	}

}
