package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/BeJloHukku/phonebook/internal/httpapi"
	"github.com/BeJloHukku/phonebook/internal/storage/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dbUrl := getDatabaseUrl()
	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		log.Fatalf("Не удалось настроить подключение: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	err = db.PingContext(ctx)
	cancel()

	if err != nil {
		log.Fatal("Не удалось подключиться к базе!")
	}

	_, err = db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS contacts (
			id          SERIAL PRIMARY KEY,
			name        TEXT NOT NULL,
			surname     TEXT,
			second_name TEXT
		);
		CREATE TABLE IF NOT EXISTS numbers (
			id         SERIAL PRIMARY KEY,
			number     TEXT NOT NULL,
			contact_id INT REFERENCES contacts(id) ON DELETE CASCADE
		);
	`)
	if err != nil {
		log.Fatalf("Не удалось выполнить миграцию: %v", err)
	}

	store := postgres.New(db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewRouter(store),
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Println("Сервер запущен на http://localhost:8080")

	if err := server.ListenAndServe(); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
