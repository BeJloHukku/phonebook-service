package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func getDatabaseUrl() string {
	if err := godotenv.Load(); err != nil &&
		!errors.Is(err, os.ErrNotExist) {
		log.Fatalf("Не удалось загрузить .env: %v", err)
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl != "" {
		return dbUrl
	}

	dbUser := os.Getenv("DATABASE_USER")
	if dbUser == "" {
		log.Fatal("Переменная DATABASE_USER не задана")
	}

	dbHost := os.Getenv("DATABASE_HOST")
	if dbHost == "" {
		log.Fatal("Переменная DATABASE_HOST не задана")
	}

	dbPassword := os.Getenv("DATABASE_PASSWORD")
	if dbPassword == "" {
		log.Fatal("Переменная DATABASE_PASSWORD не задана")
	}

	dbName := os.Getenv("DATABASE_DB")
	if dbName == "" {
		log.Fatal("Переменная DATABASE_DB не задана")
	}

	dbPort := os.Getenv("DATABASE_PORT")
	if dbPort == "" {
		log.Fatal("Переменная DATABASE_PORT не задана")
	}

	dbUrl = fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser,
		dbPassword,
		dbHost,
		dbPort,
		dbName,
	)

	return dbUrl
}
