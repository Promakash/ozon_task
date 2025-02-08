package url_shortener

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"ozon_task/domain"
	"ozon_task/internal/usecases"
	pkglog "ozon_task/pkg/log"
	urlshortenerv1 "ozon_task/protos/gen/go"
	"time"
)

type gRPCServerAPI struct {
	urlshortenerv1.UnimplementedURLShortenerServer
	service         usecases.URL
	responseTimeout time.Duration
	logger          *slog.Logger
}

func Register(gRPC *grpc.Server, URL usecases.URL, responseTimeout time.Duration, logger *slog.Logger) {
	urlshortenerv1.RegisterURLShortenerServer(gRPC, &gRPCServerAPI{
		service:         URL,
		responseTimeout: responseTimeout,
		logger:          logger,
	})
}

func (s *gRPCServerAPI) ShortenURL(
	ctx context.Context,
	request *urlshortenerv1.ShortenURLRequest,
) (*urlshortenerv1.ShortenURLResponse, error) {
	const op = "gRPCServerAPI.ShortenURL"
	log := s.logger.With(
		slog.String("op", op),
	)

	if ok := domain.IsValidOriginalURL(request.GetOriginalUrl()); !ok {
		log.Error("error while validating request", pkglog.Err(domain.ErrInvalidOriginal))
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidOriginal.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, s.responseTimeout)
	defer cancel()

	shortened, err := s.service.ShortenURL(ctx, request.GetOriginalUrl())
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
	request *urlshortenerv1.ResolveURLRequest,
) (*urlshortenerv1.ResolveURLResponse, error) {
	const op = "gRPCServerAPI.ResolveURL"
	log := s.logger.With(
		slog.String("op", op),
	)

	if ok := domain.IsValidShortenedURL(request.ShortenedUrl); !ok {
		log.Error("error while validating request", pkglog.Err(domain.ErrInvalidShortened))
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidShortened.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, s.responseTimeout)
	defer cancel()

	original, err := s.service.ResolveURL(ctx, request.ShortenedUrl)
	if err != nil {
		log.Error("failed to get original url", pkglog.Err(err))
		return nil, s.handleError(err)
	}

	return &urlshortenerv1.ResolveURLResponse{
		OriginalUrl: original,
	}, nil
}

var gRPCErrMap = map[error]error{
	context.DeadlineExceeded:    status.Error(codes.DeadlineExceeded, "deadline of operation exceeded"),
	context.Canceled:            status.Error(codes.Canceled, "operation was cancelled"),
	domain.ErrOriginalNotFound:  status.Error(codes.NotFound, domain.ErrOriginalNotFound.Error()),
	domain.ErrShortenedNotFound: status.Error(codes.NotFound, domain.ErrShortenedNotFound.Error()),
}

func (s *gRPCServerAPI) handleError(err error) error {
	if grpcErr, ok := gRPCErrMap[err]; ok {
		return grpcErr
	}
	return status.Error(codes.Internal, "internal server error")
}
