package models

type URLRecord struct {
	ShortCode string `json:"short_code" db:"short_code"`
	OriginalURL string `json:"original_url" db:"url"`
}

type CreateURLRequest struct {
	Url string `json:"url" validate:"required,http_url"`
}

type CreateURLResponse struct {
	ShortCode string `json:"short_code"`
	ShortUrl string `json:"short_url"`
	OriginalUrl string `json:"original_url"`
}

type GetURLRequest struct {
	ShortCode string `json:"short_code" validate:"required,alphanum"`
}