package app

import (
	"database/sql"
	"fmt"
	"log"
	"merch/internal/repository/pgdb"
	"merch/internal/service"
	"merch/internal/web/v1/handler"
	"merch/pkg/logger"
	"merch/pkg/postgres"
	"net/http"
	"os"
	"strconv"
)

func Run() {
	logger_ := logger.NewLogrusLogger()
	db := initDatabase()
	repo := pgdb.NewRepository(db)
	jwtSecret := getEnv("JWT_SECRET")
	service_ := service.NewService(repo, jwtSecret)
	router := handler.NewRouter(service_, logger_, jwtSecret)

	serverPort := os.Getenv("SERVER_PORT")
	if err := http.ListenAndServe(fmt.Sprintf(":%s", serverPort), router); err != nil {
		panic(err)
	}
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("missing required environment variable: %s", key)
	}
	return value
}

func getDBConfig() postgres.Config {
	port, err := strconv.Atoi(getEnv("DATABASE_PORT"))
	if err != nil {
		log.Fatalf("invalid database port: %v", err)
	}

	return postgres.Config{
		Host:     getEnv("DATABASE_HOST"),
		Port:     port,
		User:     getEnv("DATABASE_USER"),
		Password: getEnv("DATABASE_PASSWORD"),
		DBName:   getEnv("DATABASE_NAME"),
		SSLMode:  "disable",
	}
}

func initDatabase() *sql.DB {
	cfg := getDBConfig()
	db, err := postgres.New(cfg)
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}
	return db
}
