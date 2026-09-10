package links

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/CosmoS1X/go-project-278/internal/storage/sqlc"
)

const uniqueViolation = "23505"

var (
	ErrNotFound       = errors.New("link not found")
	ErrShortNameTaken = errors.New("short name already taken")
)

const (
	shortNameLength   = 8
	shortNameAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Repository interface {
	List(ctx context.Context, offset, limit int32) ([]Link, int64, error)
	GetByID(ctx context.Context, id int64) (Link, error)
	GetByShortName(ctx context.Context, shortName string) (Link, error)
	Create(ctx context.Context, originalURL, shortName string) (Link, error)
	Update(ctx context.Context, id int64, originalURL, shortName string) (Link, error)
	Delete(ctx context.Context, id int64) error
	GenerateShortName(ctx context.Context) (string, error)
	RecordVisit(ctx context.Context, linkID int64, referer, ip, userAgent string, status int32) error
	ListVisits(ctx context.Context, offset, limit int32) ([]LinkVisit, int64, error)
}

type sqlcRepository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) Repository {
	return &sqlcRepository{queries: queries}
}

func (r *sqlcRepository) List(ctx context.Context, offset, limit int32) ([]Link, int64, error) {
	rows, err := r.queries.GetLinks(ctx, sqlc.GetLinksParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("get links: %w", err)
	}

	total, err := r.queries.CountLinks(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count links: %w", err)
	}

	links := make([]Link, 0, len(rows))
	for _, row := range rows {
		links = append(links, toLink(row))
	}

	return links, total, nil
}

func (r *sqlcRepository) GetByID(ctx context.Context, id int64) (Link, error) {
	row, err := r.queries.GetLinkByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Link{}, ErrNotFound
		}
		return Link{}, fmt.Errorf("get link by id: %w", err)
	}

	return toLink(row), nil
}

func (r *sqlcRepository) Create(ctx context.Context, originalURL, shortName string) (Link, error) {
	row, err := r.queries.CreateLink(ctx, sqlc.CreateLinkParams{
		OriginalUrl: originalURL,
		ShortName:   shortName,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return Link{}, ErrShortNameTaken
		}
		return Link{}, fmt.Errorf("create link: %w", err)
	}

	return toLink(row), nil
}

func (r *sqlcRepository) Update(ctx context.Context, id int64, originalURL, shortName string) (Link, error) {
	row, err := r.queries.UpdateLink(ctx, sqlc.UpdateLinkParams{
		ID:          id,
		OriginalUrl: originalURL,
		ShortName:   shortName,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Link{}, ErrNotFound
		}
		if isUniqueViolation(err) {
			return Link{}, ErrShortNameTaken
		}
		return Link{}, fmt.Errorf("update link: %w", err)
	}

	return toLink(row), nil
}

func (r *sqlcRepository) Delete(ctx context.Context, id int64) error {
	if err := r.queries.DeleteLink(ctx, id); err != nil {
		return fmt.Errorf("delete link: %w", err)
	}

	return nil
}

func (r *sqlcRepository) GetByShortName(ctx context.Context, shortName string) (Link, error) {
	row, err := r.queries.GetLinkByShortName(ctx, shortName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Link{}, ErrNotFound
		}
		return Link{}, fmt.Errorf("get link by short name: %w", err)
	}

	return toLink(row), nil
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

func (r *sqlcRepository) GenerateShortName(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	for {
		buf := make([]byte, shortNameLength)
		for i := range buf {
			n, err := rand.Int(rand.Reader, big.NewInt(int64(len(shortNameAlphabet))))
			if err != nil {
				return "", err
			}
			buf[i] = shortNameAlphabet[n.Int64()]
		}

		candidate := string(buf)
		_, err := r.queries.GetLinkByShortName(ctx, candidate)
		if errors.Is(err, sql.ErrNoRows) {
			return candidate, nil
		}
		if err != nil {
			return "", err
		}
	}
}

func toLink(row sqlc.Link) Link {
	return Link{
		ID:          row.ID,
		OriginalURL: row.OriginalUrl,
		ShortName:   row.ShortName,
		CreatedAt:   row.CreatedAt,
	}
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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}
