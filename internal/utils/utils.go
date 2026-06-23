package utils

import (
	"math/rand/v2"
	"os"
	"strings"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortCode(length ...int) string {
	// Set default value
	n := 10
	if len(length) > 0 {
		n = length[0]
	}

	var sb strings.Builder
	sb.Grow(n)

	for i := 0; i < n; i++ {
		// rand.IntN is faster and safer in Go 1.22+
		sb.WriteByte(charset[rand.IntN(len(charset))])
	}

	return sb.String()
}

func ConstructShortURL(shortCode string) string {
	return os.Getenv("BASE_URL") + "/urls/" + shortCode
}