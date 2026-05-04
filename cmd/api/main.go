package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/pratikdev/url-shortner-api/internal/database"
)

func main() {
    handler := slog.NewJSONHandler(os.Stderr, nil)
    logger := slog.New(handler)
	pool := database.InitDB(logger)

    defer pool.Close() // Close the database connection pool when the application exits

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.Write([]byte(`{"status":"ok"}`))
    })
    http.ListenAndServe(":" + os.Getenv("API_PORT"), nil)
}