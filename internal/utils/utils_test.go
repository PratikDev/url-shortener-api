package utils

import (
	"bytes"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGenerateShortCode(t *testing.T) {
	generatedShortCode := GenerateShortCode()

	// checks if short-code default length matches
	t.Run("default length", func(t *testing.T) {
		if len(generatedShortCode) != SHORT_CODE_DEFAULT_LENGTH {
			t.Errorf("expected: %d, got: %d", SHORT_CODE_DEFAULT_LENGTH, len(generatedShortCode))
		}
	})

	// checks if short-code is of the character set
	t.Run("uses only charset characters", func(t *testing.T) {
		for _, c := range generatedShortCode {
			if !strings.ContainsRune(CHARSET, c) {
				t.Errorf("character %q not in charset", c)
			}
		}
	})

	// checks if custom length in shortcode works
	t.Run("custom length", func(t *testing.T) {
		customLength := 20
		generatedShortCodeCustomLength := GenerateShortCode(customLength)
		if len(generatedShortCodeCustomLength) != customLength {
			t.Errorf("expected: %d, got: %d", customLength, len(generatedShortCodeCustomLength))
		}
	})

	// checks if two generated short-codes doesn't match
	t.Run("produces different codes", func(t *testing.T) {
		generatedShortCodeA := GenerateShortCode()
		generatedShortCodeB := GenerateShortCode()
		if generatedShortCodeA == generatedShortCodeB {
			t.Errorf("expected: %s, got: %s", "Not match", "Match")
		}
	})
}

func TestGetClientIP(t *testing.T) {
	bodyReader := bytes.NewReader([]byte{})
	IP := "125.0.0.1"

	t.Run("X-Forwarded-For", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", bodyReader)
		req.Header.Set("X-Forwarded-For", IP)
		clientIP := GetClientIP(req)
		if clientIP != IP {
			t.Errorf("expected: %s, got: %s", IP, clientIP)
		}
	})

	t.Run("X-Real-IP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", bodyReader)
		req.Header.Set("X-Real-IP", IP)
		clientIP := GetClientIP(req)
		if clientIP != IP {
			t.Errorf("expected: %s, got: %s", IP, clientIP)
		}
	})

	t.Run("no headers IP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", bodyReader)
		clientIP := GetClientIP(req)
		expectedIP, _, _ := net.SplitHostPort(req.RemoteAddr)
		if clientIP != expectedIP {
			t.Errorf("expected: %s, got: %s", expectedIP, clientIP)
		}
	})

	t.Run("malformed IP", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", bodyReader)
		req.Header.Set("X-Real-IP", "gibbrish")
		clientIP := GetClientIP(req)
		expectedIP, _, _ := net.SplitHostPort(req.RemoteAddr)
		if clientIP != expectedIP {
			t.Errorf("expected: %s, got: %s", expectedIP, clientIP)
		}
	})
}