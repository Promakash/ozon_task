package url_shortener

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ozon_task/domain"
	"ozon_task/internal/usecases"
	urlshortenerv1 "ozon_task/protos/gen/go"
	"time"
)

type gRPCURLService struct {
	urlshortenerv1.UnimplementedURLShortenerServer
	service         usecases.URL
	responseTimeout time.Duration
}

func Register(gRPC *grpc.Server, URL usecases.URL) {
	urlshortenerv1.RegisterURLShortenerServer(gRPC, &gRPCURLService{service: URL})
}

func (s *gRPCURLService) ShortenURL(ctx context.Context, request *urlshortenerv1.ShortenURLRequest) (*urlshortenerv1.ShortenURLResponse, error) {
	if ok := domain.IsValidOriginalURL(request.GetOriginalUrl()); !ok {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidOriginal.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, s.responseTimeout)
	defer cancel()

	shortened, err := s.service.ShortenURL(ctx, request.GetOriginalUrl())
	if err != nil {
		return nil, s.handleError(err)
	}

	return &urlshortenerv1.ShortenURLResponse{
		ShortenedUrl: shortened,
	}, nil
}

func (s *gRPCURLService) ResolveURL(ctx context.Context, request *urlshortenerv1.ResolveURLRequest) (*urlshortenerv1.ResolveURLResponse, error) {
	if ok := domain.IsValidShortenedURL(request.ShortenedUrl); !ok {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidShortened.Error())
	}

	ctx, cancel := context.WithTimeout(ctx, s.responseTimeout)
	defer cancel()

	original, err := s.service.ResolveURL(ctx, request.ShortenedUrl)
	if err != nil {
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

func (s *gRPCURLService) handleError(err error) error {
	if grpcErr, ok := gRPCErrMap[err]; ok {
		return grpcErr
	}
	return status.Error(codes.Internal, "internal server error")
}
