package url_shortener

import (
	"context"
	"errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"ozon_task/domain"
	"ozon_task/internal/usecases"
	pkgerr "ozon_task/pkg/error"
	pkglog "ozon_task/pkg/log"
	urlshortenerv1 "ozon_task/protos/gen/go"
	"time"
)

type gRPCServerAPI struct {
	urlshortenerv1.UnimplementedURLShortenerServer
	service           usecases.URL
	operationsTimeout time.Duration
	logger            *slog.Logger
}

func Register(gRPC *grpc.Server, URL usecases.URL, operationsTimeout time.Duration, logger *slog.Logger) {
	urlshortenerv1.RegisterURLShortenerServer(gRPC, &gRPCServerAPI{
		service:           URL,
		operationsTimeout: operationsTimeout,
		logger:            logger,
	})
}

func (s *gRPCServerAPI) ShortenURL(
	ctx context.Context,
	req *urlshortenerv1.ShortenURLRequest,
) (*urlshortenerv1.ShortenURLResponse, error) {
	const op = "gRPCServerAPI.ShortenURL"
	log := s.logger.With(
		slog.String("op", op),
	)

	originalURL := domain.NormalizeURL(req.GetOriginalUrl())

	if ok, err := domain.IsValidOriginalURL(originalURL); !ok {
		log.Error("error while validating req", pkglog.Err(err))
		return nil, s.handleError(err)
	}

	ctx, cancel := context.WithTimeout(ctx, s.operationsTimeout)
	defer cancel()

	shortened, err := s.service.ShortenURL(ctx, originalURL)
	if err != nil {
		log.Error("failed to generate shortened url", pkglog.Err(err))
		return nil, s.handleError(err)
	}

	return &urlshortenerv1.ShortenURLResponse{
		ShortenedUrl: shortened,
	}, nil
}

func (s *gRPCServerAPI) ResolveURL(
	ctx context.Context,
	req *urlshortenerv1.ResolveURLRequest,
) (*urlshortenerv1.ResolveURLResponse, error) {
	const op = "gRPCServerAPI.ResolveURL"
	log := s.logger.With(
		slog.String("op", op),
	)

	if ok, err := domain.IsValidShortenedURL(req.GetShortenedUrl()); !ok {
		log.Error("error while validating req", pkglog.Err(err))
		return nil, s.handleError(err)
	}

	ctx, cancel := context.WithTimeout(ctx, s.operationsTimeout)
	defer cancel()

	original, err := s.service.ResolveURL(ctx, req.GetShortenedUrl())
	if err != nil {
		log.Error("failed to get original url", pkglog.Err(err))
		return nil, s.handleError(err)
	}

	return &urlshortenerv1.ResolveURLResponse{
		OriginalUrl: original,
	}, nil
}

func (s *gRPCServerAPI) handleError(err error) error {
	err = pkgerr.UnwrapAll(err)

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "deadline of operation exceeded")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "operation was cancelled")
	case errors.Is(err, domain.ErrOriginalNotFound),
		errors.Is(err, domain.ErrShortenedNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInaccessibleOriginal),
		errors.Is(err, domain.ErrInvalidOriginal),
		errors.Is(err, domain.ErrInvalidShortened):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
