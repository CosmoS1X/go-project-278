package links

import (
	"strings"
	"time"
)

type Link struct {
	ID          int64
	OriginalURL string
	ShortName   string
	CreatedAt   time.Time
}

type request struct {
	OriginalURL string `json:"original_url" binding:"required,http_url"`
	ShortName   string `json:"short_name"   binding:"omitempty,min=3,max=32"`
}

func (r *request) Normalize() {
	r.OriginalURL = strings.TrimSpace(r.OriginalURL)
	r.ShortName = strings.TrimSpace(r.ShortName)
}

type response struct {
	ID          int64     `json:"id"`
	OriginalURL string    `json:"original_url"`
	ShortName   string    `json:"short_name"`
	ShortURL    string    `json:"short_url"`
	CreatedAt   time.Time `json:"created_at"`
}
