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
