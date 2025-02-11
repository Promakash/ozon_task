package suite

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"ozon_task/internal/config"
	urlshortenerv1 "ozon_task/protos/gen/go"
	"testing"
	"time"
)

const hostEnvVar = "GRPC_HOST"

var gRPCHost = os.Getenv(hostEnvVar)

type Suite struct {
	*testing.T
	Cfg       *config.Config
	URLClient urlshortenerv1.URLShortenerClient
}

func New(t *testing.T) (context.Context, *Suite) {
	t.Helper()
	t.Parallel()

	const OperationsTimeout = time.Second * 10

	ctx, cancel := context.WithTimeout(context.Background(), OperationsTimeout)

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
		URLClient: urlshortenerv1.NewURLShortenerClient(cc),
	}
}
