package utils

import (
	"math/rand/v2"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const CHARSET = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const SHORT_CODE_DEFAULT_LENGTH = 10

func GenerateShortCode(length ...int) string {
	// Set default value
	n := SHORT_CODE_DEFAULT_LENGTH
	if len(length) > 0 {
		n = length[0]
	}

	var sb strings.Builder
	sb.Grow(n)

	for i := 0; i < n; i++ {
		// rand.IntN is faster and safer in Go 1.22+
		sb.WriteByte(CHARSET[rand.IntN(len(CHARSET))])
	}

	return sb.String()
}

func ConstructShortURL(shortCode string) string {
	return os.Getenv("BASE_URL") + "/urls/" + shortCode
}

func FloatToString(val float64) string {
	return strconv.FormatFloat(val, 'f', -1, 64)
}

func GetClientIP(r *http.Request) string {
	// 1. Check standard proxy headers
	headers := []string{"X-Forwarded-For", "X-Real-IP"}
	for _, header := range headers {
		addresses := r.Header.Get(header)
		if addresses != "" {
			// X-Forwarded-For can contain a comma-separated list of IPs.
			// The first IP is usually the original client.
			for ip := range strings.SplitSeq(addresses, ",") {
				ip = strings.TrimSpace(ip)
				// Handle potential IPv6 formatting wrapped in brackets
				if strings.HasPrefix(ip, "[") && strings.HasSuffix(ip, "]") {
					ip = ip[1 : len(ip)-1]
				}
				if net.ParseIP(ip) != nil {
					return ip
				}
			}
		}
	}

	// 2. Fallback to RemoteAddr (Format is always "IP:port" or "[IPv6]:port")
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If SplitHostPort fails (e.g. no port present), return RemoteAddr as-is
		return r.RemoteAddr
	}
	return ip
}