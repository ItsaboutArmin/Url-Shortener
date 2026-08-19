package repository

import (
	"context"
	"database/sql"
	"errors"

	"URL-Shotener/internal/model"
)

type URLRepository interface {
	Save(ctx context.Context, url *model.URL) error
	FindByShortCode(ctx context.Context, shortCode string) (*model.URL, error)
}

type postgresURLRepo struct {
	db *sql.DB
}

func NewPostgresURLRepository(db *sql.DB) URLRepository {
	return &postgresURLRepo{db: db}
}

func (r *postgresURLRepo) Save(ctx context.Context, url *model.URL) error {
	query := `
		INSERT INTO urls (original_url, short_code, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query,
		url.OriginalURL, url.ShortCode, url.ExpiresAt,
	).Scan(&url.ID, &url.CreatedAt)

	if err != nil {
		return err
	}
	return nil
}

func (r *postgresURLRepo) FindByShortCode(ctx context.Context, shortCode string) (*model.URL, error) {
	query := `
		SELECT id, original_url, short_code, created_at, expires_at
		FROM urls
		WHERE short_code = $1
	`
	var url model.URL
	err := r.db.QueryRowContext(ctx, query, shortCode).Scan(
		&url.ID, &url.OriginalURL, &url.ShortCode, &url.CreatedAt, &url.ExpiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &url, nil
}
