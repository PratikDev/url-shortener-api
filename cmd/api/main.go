package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/pratikdev/url-shortener-api/internal/database"
	"github.com/pratikdev/url-shortener-api/internal/handlers"
	"github.com/pratikdev/url-shortener-api/internal/middleware"
)

func main() {
    handler := slog.NewJSONHandler(os.Stderr, nil)
    logger := slog.New(handler)
	pool := database.InitDB(logger)
    validate := validator.New()
    mux := http.NewServeMux()

    defer pool.Close() // Close the database connection pool when the application exits

    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        handlers.GetHealth(w, r)
    })
    mux.HandleFunc("POST /urls", func(w http.ResponseWriter, r *http.Request) {
        handlers.CreateURL(w, r, pool, validate)
    })
    mux.HandleFunc("GET /urls/{shortCode}", func(w http.ResponseWriter, r *http.Request) {
        handlers.GetURL(w, r, pool, validate)
    })

    http.ListenAndServe(":" + os.Getenv("PORT"), middleware.LoggingMiddleware(
        middleware.RateLimitMiddleware(mux),
        logger,
    ))
}