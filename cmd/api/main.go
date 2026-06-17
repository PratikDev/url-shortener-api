package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/pratikdev/url-shortener-api/internal/database"
	"github.com/pratikdev/url-shortener-api/internal/handlers"
)

func main() {
    handler := slog.NewJSONHandler(os.Stderr, nil)
    logger := slog.New(handler)
	pool := database.InitDB(logger)
    validate := validator.New()

    defer pool.Close() // Close the database connection pool when the application exits

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    })
    http.HandleFunc("POST /urls", func(w http.ResponseWriter, r *http.Request) {
        handlers.CreateURL(w, r, pool, validate)
    })
    http.HandleFunc("GET /urls/{shortCode}", func(w http.ResponseWriter, r *http.Request) {
        handlers.GetURL(w, r, pool, validate)
    })

    http.ListenAndServe(":" + os.Getenv("API_PORT"), nil)
}