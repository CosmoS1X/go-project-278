package visits

import (
	"context"
	"fmt"

	"github.com/CosmoS1X/go-project-278/internal/storage/sqlc"
)

type Repository interface {
	RecordVisit(ctx context.Context, linkID int64, referer, ip, userAgent string, status int32) error
	ListVisits(ctx context.Context, offset, limit int32) ([]LinkVisit, int64, error)
}

type sqlcRepository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) Repository {
	return &sqlcRepository{queries: queries}
}

func (r *sqlcRepository) RecordVisit(ctx context.Context, linkID int64, referer, ip, userAgent string, status int32) error {
	if _, err := r.queries.CreateVisit(ctx, sqlc.CreateVisitParams{
		LinkID:    linkID,
		Ip:        ip,
		Referer:   referer,
		UserAgent: userAgent,
		Status:    status,
	}); err != nil {
		return fmt.Errorf("record visit: %w", err)
	}

	return nil
}

func (r *sqlcRepository) ListVisits(ctx context.Context, offset, limit int32) ([]LinkVisit, int64, error) {
	rows, err := r.queries.GetLinkVisits(ctx, sqlc.GetLinkVisitsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("get link visits: %w", err)
	}

	total, err := r.queries.CountLinkVisits(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count link visits: %w", err)
	}

	visits := make([]LinkVisit, 0, len(rows))
	for i := range rows {
		visits = append(visits, toLinkVisit(&rows[i]))
	}

	return visits, total, nil
}

func toLinkVisit(row *sqlc.LinkVisit) LinkVisit {
	return LinkVisit{
		ID:        row.ID,
		LinkID:    row.LinkID,
		IP:        row.Ip,
		UserAgent: row.UserAgent,
		Status:    row.Status,
		Referer:   row.Referer,
		CreatedAt: row.CreatedAt,
	}
}
