package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pratikdev/url-shortener-api/internal/database"
	"github.com/pratikdev/url-shortener-api/internal/models"
	"github.com/pratikdev/url-shortener-api/internal/testutils"
)

func TestCreateURL(t *testing.T) {
	// truncate the table first
	testutils.TruncateURLsTable(testPool)

	// create payload
	url := "https://example.com"
	payload := models.CreateURLRequest{
		Url: url,
	}
	jsonBytes, _ := json.Marshal(payload)

	// configure request and recorder
	r := httptest.NewRequest(http.MethodPost, "/urls", bytes.NewBuffer(jsonBytes))
	w := httptest.NewRecorder()

	// ping
	CreateURL(w, r, testPool, testValidator)

	var responseBody models.CreateURLResponse
	
	t.Run("status check", func(t *testing.T) {
		expectedCode := 201
		if w.Code != expectedCode {
			t.Errorf("wrong status code. expected: %d, got: %d", expectedCode, w.Code)
		}
	})

	t.Run("response body shape", func(t *testing.T) {
		// read response body
		err := json.NewDecoder(w.Body).Decode(&responseBody)
		if err != nil {
			t.Errorf("wrong response body shape")
		}
	})

	t.Run("field correctness", func(t *testing.T) {
		if responseBody == (models.CreateURLResponse{}) {
			t.Errorf("empty response body")
			return
		}

		if responseBody.OriginalUrl != url {
			t.Errorf("wrong original url in response body")
		}

		if len(responseBody.ShortCode) == 0 {
			t.Errorf("empty short-code in response body")
		}

		if !strings.Contains(responseBody.ShortUrl, responseBody.ShortCode) {
			t.Errorf("short-code is not in shortUrl in response body")
		}
	})

	t.Run("database sideeffect", func(t *testing.T) {
		if responseBody == (models.CreateURLResponse{}) {
			t.Errorf("empty response body")
			return
		}

		urlRecord, err := database.GetURLFromShortCode(testPool, responseBody.ShortCode)
		if err != nil {
			t.Errorf("failed querying url record from response body short-code")
		}

		if urlRecord.OriginalURL != responseBody.OriginalUrl {
			t.Errorf("orginal-url mismatch with db and response body")
		}

		if urlRecord.ShortCode != responseBody.ShortCode {
			t.Errorf("short-code mismatch with db and response body")
		}
	})
}