package database

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pratikdev/url-shortener-api/internal/models"
	"github.com/pratikdev/url-shortener-api/internal/utils"
)

var ErrDuplicateShortCode = errors.New("duplicate short code")

func InsertNewURL(pool *pgxpool.Pool, Url string) (models.URLRecord, error) {
	MAX_RETRIES := 5
	for range MAX_RETRIES {
		short_code := utils.GenerateShortCode()

		query := `INSERT INTO urls (short_code, url) VALUES (@short_code, @url) RETURNING short_code, url`
		args := pgx.NamedArgs{
			"short_code": short_code,
			"url": Url,
		}
		rows, err := pool.Query(context.Background(), query, args)
		
		if err != nil {
			// if duplicate short code
			if err, ok := err.(*pgconn.PgError); ok && err.Code == pgerrcode.UniqueViolation {
				continue
			}

			// any other error
			return models.URLRecord{}, err
		}
		
		// if success
		return pgx.CollectOneRow(rows, pgx.RowToStructByName[models.URLRecord])
	}

	return models.URLRecord{}, ErrDuplicateShortCode
}