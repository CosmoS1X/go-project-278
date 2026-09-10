package links

import (
	"time"
)

type Link struct {
	ID          int64
	OriginalURL string
	ShortName   string
	CreatedAt   time.Time
}

type request struct {
	OriginalURL string `json:"original_url"`
	ShortName   string `json:"short_name"`
}

type response struct {
	ID          int64     `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortName   string    `json:"short_name"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
}
