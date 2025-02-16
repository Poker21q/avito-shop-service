package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

var (
	ErrConnectionFailed = errors.New("postgres: connection failed")
)

func New(cfg Config) (*sql.DB, error) {
	dsn := createDataSourceName(cfg)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, ErrConnectionFailed
	}

	if err = pingDatabase(db); err != nil {
		return nil, ErrConnectionFailed
	}

	return db, nil
}

func createDataSourceName(cfg Config) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
}

func pingDatabase(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
