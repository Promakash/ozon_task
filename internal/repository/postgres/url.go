package postgres

import (
	"context"
	"errors"

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

func (r *URLRepository) PutShortenedURL(ctx context.Context, original domain.URL, shortened *domain.URL) error {
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

func (r *URLRepository) GetOriginalURLByShortened(ctx context.Context, shortened domain.URL) (domain.URL, error) {
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
		return "", err
	}

	return original, nil
}

func (r *URLRepository) GetShortenedURLByOriginal(ctx context.Context, original domain.URL) (domain.URL, error) {
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
		return "", err
	}

	return shortened, nil
}
