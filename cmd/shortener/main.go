package main

import (
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kirill555525/URL/cmd/config"
	"github.com/kirill555525/URL/internal/controllers"
	"github.com/kirill555525/URL/internal/logger"
	"github.com/kirill555525/URL/internal/store"
	"github.com/kirill555525/URL/internal/store/file"
	"github.com/kirill555525/URL/internal/store/memory"
	"github.com/kirill555525/URL/internal/store/postgres"
	"net/http"
)

func main() {
	cfg := config.Init()
	err := logger.Initialize(cfg.FlagLogLevel)
	if err != nil {
		panic(err)
	}

	var storage store.Store

	if cfg.DatabaseDSN != "" {
		storage, err = postgres.NewStore(cfg.DatabaseDSN)
		if err != nil {
			panic(err)
		}
	} else {
		storage = memory.NewStore()
		if cfg.FileStoragePath != "" {
			storage, err = file.NewStore(cfg.FileStoragePath, storage)
			if err != nil {
				panic(err)
			}
		}
	}

	defer storage.Close()

	controller := controllers.NewBaseController(storage)
	router := chi.NewRouter()
	router.Mount("/", controller.Route())

	if err := http.ListenAndServe(cfg.Addr, router); err != nil {
		panic(err)
	}
}
