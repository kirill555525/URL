package database

import (
	"context"
	"database/sql"
	"github.com/kirill555525/URL/cmd/config"
	"time"
)

var db *sql.DB

func ConnectDB() (*sql.DB, error) {
	cfg := config.GetConfig()

	dataBase, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	db = dataBase

	return db, nil
}

func GetDB() *sql.DB {
	return db
}

func CheckConnectDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}

	return nil

}
