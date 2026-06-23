package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pratikdev/url-shortener-api/internal/database"
	"github.com/pratikdev/url-shortener-api/internal/models"
	"github.com/pratikdev/url-shortener-api/internal/utils"
)

func GetHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

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

func GetURL(w http.ResponseWriter, r *http.Request, pool *pgxpool.Pool, validate *validator.Validate) {
	request_body := models.GetURLRequest{
		ShortCode: r.PathValue("shortCode"),
	}

	// validate the request body
	err := validate.Struct(request_body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// get the url
	record, err := database.GetURLFromShortCode(pool, request_body.ShortCode)
	if err != nil {
		// if no row err
		if errors.Is(err, database.ErrShortCodeNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, record.OriginalURL, http.StatusFound)
}