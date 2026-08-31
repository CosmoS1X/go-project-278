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
	List(ctx context.Context) ([]Link, error)
	GetByID(ctx context.Context, id int64) (Link, error)
	Create(ctx context.Context, originalURL, shortName string) (Link, error)
	Update(ctx context.Context, id int64, originalURL, shortName string) (Link, error)
	Delete(ctx context.Context, id int64) error
	GenerateShortName(ctx context.Context) (string, error)
}

type sqlcRepository struct {
	queries *sqlc.Queries
}

func NewRepository(queries *sqlc.Queries) Repository {
	return &sqlcRepository{queries: queries}
}

func (r *sqlcRepository) List(ctx context.Context) ([]Link, error) {
	rows, err := r.queries.GetLinks(ctx)
	if err != nil {
		return nil, fmt.Errorf("get links: %w", err)
	}

	links := make([]Link, 0, len(rows))
	for _, row := range rows {
		links = append(links, toLink(row))
	}

	return links, nil
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

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation
}
