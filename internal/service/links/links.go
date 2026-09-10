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

type LinkVisit struct {
	ID        int64
	LinkID    int64
	IP        string
	UserAgent string
	Status    int32
	Referer   string
	CreatedAt time.Time
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

type linkVisitResponse struct {
	ID        int64  `json:"id"`
	LinkID    int64  `json:"link_id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Status    int32  `json:"status"`
	// reffer (not referer) matches the frozen frontend bundle; DB column is referer.
	Referer   string    `json:"reffer"`
	CreatedAt time.Time `json:"created_at"`
}

func toLinkVisitResponse(v *LinkVisit) linkVisitResponse {
	return linkVisitResponse(*v)
}
