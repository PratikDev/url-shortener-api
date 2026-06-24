package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/pratikdev/url-shortener-api/internal/utils"
)

func LoggingMiddleware(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		clientIp := utils.GetClientIP(r)
		wrappedWriter := utils.NewResponseWrapper(w)

		next.ServeHTTP(wrappedWriter, r)
		
		duration := time.Since(start)
		logger.Info("request handled",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrappedWriter.StatusCode,
			"duration_ms", duration.Milliseconds(),
			"client_ip", clientIp,
		)
	})
}