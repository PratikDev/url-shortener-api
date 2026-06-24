package middleware

import (
	"fmt"
	"net/http"

	"github.com/pratikdev/url-shortener-api/internal/models"
	"github.com/pratikdev/url-shortener-api/internal/utils"
)

var MAX_TOKENS = 100
var REFILL_RATE = 5

func RateLimitMiddleware (next http.Handler) http.Handler {
	safemap := utils.NewSafeMap[models.TokenBucket]()

	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// rate limit logic goes here
		client_ip := utils.GetClientIP(r)

		fmt.Printf(client_ip, safemap)

		next.ServeHTTP(w, r)
	})
}