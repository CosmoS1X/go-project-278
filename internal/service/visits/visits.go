package visits

import "time"

type LinkVisit struct {
	ID        int64
	LinkID    int64
	IP        string
	UserAgent string
	Status    int32
	Referer   string
	CreatedAt time.Time
}

type response struct {
	ID        int64  `json:"id"`
	LinkID    int64  `json:"link_id"`
	IP        string `json:"ip"`
	UserAgent string `json:"user_agent"`
	Status    int32  `json:"status"`
	// reffer (not referer) matches the frozen frontend bundle; DB column is referer.
	Referer   string    `json:"reffer"`
	CreatedAt time.Time `json:"created_at"`
}

func toResponse(v *LinkVisit) response {
	return response(*v)
}
