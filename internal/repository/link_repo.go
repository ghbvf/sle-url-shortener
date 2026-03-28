package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

var ErrCodeExists = errors.New("code already exists")

type Link struct {
	ID        int64
	Code      string
	URL       string
	Clicks    int64
	CreatedAt time.Time
	ExpiresAt *time.Time
}

type LinkRepository interface {
	Create(ctx context.Context, link *Link) error
}

type PostgresLinkRepo struct {
	db *sql.DB
}

func NewPostgresLinkRepo(db *sql.DB) *PostgresLinkRepo {
	return &PostgresLinkRepo{db: db}
}

// Create inserts a new link. Returns ErrCodeExists on duplicate code.
// Consistency: L1 — single table insert, no cross-service side effects.
func (r *PostgresLinkRepo) Create(ctx context.Context, link *Link) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO links (code, url) VALUES ($1, $2)`,
		link.Code, link.URL,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrCodeExists
		}
		return fmt.Errorf("insert link: %w", err)
	}
	return nil
}
