package url_shortener

import (
	"context"
	"google.golang.org/grpc"
	"ozon_task/internal/usecases"
	urlshortenerv1 "ozon_task/protos/gen/go"
)

type gRPCURLService struct {
	urlshortenerv1.UnimplementedURLShortenerServer
	URL usecases.URL
}

func Register(gRPC *grpc.Server, URL usecases.URL) {
	urlshortenerv1.RegisterURLShortenerServer(gRPC, &gRPCURLService{URL: URL})
}

func (s *gRPCURLService) ShortenURL(ctx context.Context, request *urlshortenerv1.ShortenURLRequest) (*urlshortenerv1.ShortenURLResponse, error) {
	return &urlshortenerv1.ShortenURLResponse{
		ShortenedUrl: "",
	}, nil
}

func (s *gRPCURLService) ResolveURL(ctx context.Context, request *urlshortenerv1.ResolveURLRequest) (*urlshortenerv1.ResolveURLResponse, error) {
	return &urlshortenerv1.ResolveURLResponse{
		OriginalUrl: "",
	}, nil
}
