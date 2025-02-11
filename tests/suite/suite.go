package suite

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"ozon_task/internal/config"
	pkgconfig "ozon_task/pkg/config"
	urlshortenerv1 "ozon_task/protos/gen/go"
	"testing"
)

const configEnvVar = "SHORTENER_CONFIG"
const gRPCHost = "localhost:5050"

type Suite struct {
	*testing.T
	Cfg       *config.Config
	URLClient urlshortenerv1.URLShortenerClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	cfg := config.Config{}
	pkgconfig.MustLoad(configEnvVar, &cfg)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.GRPC.OperationsTimeout)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	cc, err := grpc.NewClient(
		gRPCHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc server connection failed: %v", err)
	}

	return ctx, &Suite{
		T:         t,
		Cfg:       &cfg,
		URLClient: urlshortenerv1.NewURLShortenerClient(cc),
	}
}
