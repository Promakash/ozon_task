package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ozon_task/domain"
	"ozon_task/internal/repository"
)

type URLRepository struct {
	pool *pgxpool.Pool
}

func NewURLRepository(pool *pgxpool.Pool) repository.URL {
	return &URLRepository{
		pool: pool,
	}
}

func (r *URLRepository) PutShortenedURL(ctx context.Context, original domain.URL, shortened *domain.ShortURL) error {
	var result domain.URL

	query := `
        INSERT INTO links (original_link, shortened_link)
        VALUES ($1, $2)
        ON CONFLICT (original_link)
        DO NOTHING
        RETURNING shortened_link
    `

	err := r.pool.QueryRow(ctx, query, original, shortened).Scan(&result)
	if errors.Is(err, pgx.ErrNoRows) {
		*shortened = result
		return nil
	}

	return err
}

func (r *URLRepository) GetOriginalURLByShortened(ctx context.Context, shortened domain.ShortURL) (domain.URL, error) {
	var original domain.URL
	query := `
        SELECT original_link FROM links
        WHERE shortened_link = $1
    `

	err := r.pool.QueryRow(ctx, query, shortened).Scan(&original)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrOriginalNotFound
		}
		return "", fmt.Errorf("GetOriginalURLByShortened: query failed: %w", err)
	}

	return original, nil
}

func (r *URLRepository) GetShortenedURLByOriginal(ctx context.Context, original domain.URL) (domain.ShortURL, error) {
	var shortened domain.URL
	query := `
        SELECT shortened_link FROM links
        WHERE original_link = $1
    `

	err := r.pool.QueryRow(ctx, query, original).Scan(&shortened)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", domain.ErrShortenedNotFound
		}
		return "", fmt.Errorf("GetShortenedURLByOriginal: query failed: %w", err)
	}

	return shortened, nil
}
