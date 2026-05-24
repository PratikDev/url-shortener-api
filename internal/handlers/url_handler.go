package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pratikdev/url-shortener-api/internal/database"
	"github.com/pratikdev/url-shortener-api/internal/models"
	"github.com/pratikdev/url-shortener-api/internal/utils"
)

func CreateURL(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool, validate *validator.Validate) {
	var url_request_body models.CreateURLRequest

	// Read the request body
	err := json.NewDecoder(r.Body).Decode(&url_request_body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// validate the request body
	err = validate.Struct(url_request_body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// insert the url
	record, err := database.InsertNewURL(pool, url_request_body.Url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// construct the response
	response := models.CreateURLResponse{
		ShortCode: record.ShortCode,
		OriginalUrl: record.OriginalURL,
		ShortUrl: utils.ConstructShortURL(record.ShortCode),
	}

	// convert to json
	jsonData, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(jsonData)
}