package middleware

import (
	"net/http"
	"time"

	"github.com/pratikdev/url-shortener-api/internal/models"
	"github.com/pratikdev/url-shortener-api/internal/utils"
)

var MAX_TOKENS = float64(100)
var REFILL_RATE_PER_SECOND = float64(5) // per second

func RateLimitMiddleware (next http.Handler) http.Handler {
	safemap := utils.NewSafeMap[models.TokenBucket]()
	
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		client_ip := utils.GetClientIP(r)
		allowed := false

		newBucket := safemap.Update(client_ip, func(current models.TokenBucket, exists bool) models.TokenBucket {
			if !exists {
				allowed = true
				return models.TokenBucket{Tokens: MAX_TOKENS - 1, LastRefill: time.Now()}
			}
			elapsed := time.Since(current.LastRefill)
			refilled := min(MAX_TOKENS, current.Tokens+elapsed.Seconds() * REFILL_RATE_PER_SECOND)
			if refilled >= 1 {
				allowed = true
				refilled -= 1
			}
			return models.TokenBucket{Tokens: refilled, LastRefill: time.Now()}
		})

		w.Header().Set("RateLimit-Limit", utils.FloatToString(MAX_TOKENS))
		w.Header().Set("RateLimit-Remaining", utils.FloatToString(newBucket.Tokens))
		
		tokenResetTime := (MAX_TOKENS - newBucket.Tokens) / REFILL_RATE_PER_SECOND
		w.Header().Set("RateLimit-Reset", utils.FloatToString(tokenResetTime))

		if allowed {
			next.ServeHTTP(w, r)
			return
		}

		retryAfter := (1 - newBucket.Tokens) / REFILL_RATE_PER_SECOND
		w.Header().Set("Retry-After", utils.FloatToString(retryAfter))

		// reject the request
		http.Error(w, "Error", http.StatusTooManyRequests)
	})
}