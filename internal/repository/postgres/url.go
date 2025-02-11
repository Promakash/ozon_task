package postgres

import (
	"context"
	"errors"
	"fmt"
	"ozon_task/pkg/infra/cache"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"ozon_task/domain"
	"ozon_task/internal/repository"
)

type URLRepository struct {
	pool              *pgxpool.Pool
	cache             cache.Cache
	cacheTTL          time.Duration
	cacheWriteTimeout time.Duration
}

func NewURLRepository(pool *pgxpool.Pool, cache cache.Cache, cacheTTL, cacheWriteTimeout time.Duration) repository.URL {
	return &URLRepository{
		pool:              pool,
		cache:             cache,
		cacheTTL:          cacheTTL,
		cacheWriteTimeout: cacheWriteTimeout,
	}
}

func (r *URLRepository) CreateOrGetShortenedURL(ctx context.Context, original domain.URL, shortened domain.ShortURL) (domain.ShortURL, error) {
	var result domain.URL

	// used update on in case of concurrent inserting
	query := `
        INSERT INTO links (original_link, shortened_link)
		VALUES ($1, $2)
		ON CONFLICT (original_link)
		DO UPDATE SET shortened_link = links.shortened_link
		RETURNING shortened_link;
    `

	err := r.pool.QueryRow(ctx, query, original, shortened).Scan(&result)
	if err != nil {
		return "", nil
	}

	go r.cacheURLs(original, shortened)

	return result, nil
}

func (r *URLRepository) GetOriginalURLByShortened(ctx context.Context, shortened domain.ShortURL) (domain.URL, error) {
	var original domain.URL
	if err := r.cache.Get(ctx, shortened, &original); err == nil {
		return original, nil
	}

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
	var shortened domain.ShortURL
	if err := r.cache.Get(ctx, original, &shortened); err == nil {
		return shortened, nil
	}

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

func (r *URLRepository) cacheURLs(original domain.URL, shortened domain.ShortURL) {
	// r.cacheWriteTimeout*2 because we have two write operations
	ctx, cancel := context.WithTimeout(context.Background(), r.cacheWriteTimeout*2)
	defer cancel()
	_ = r.cache.Set(ctx, original, shortened, r.cacheTTL)
	_ = r.cache.Set(ctx, shortened, original, r.cacheTTL)
}
