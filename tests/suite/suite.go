package suite

import (
	"context"
	"fmt"
	"net/http"
	"os"
	urlshortenerv1 "ozon_task/protos/gen/go"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	grpcHostEnvVar = "GRPC_HOST"
	httpHostEnvVar = "HTTP_HOST"
)

var (
	grpcHost = os.Getenv(grpcHostEnvVar)
	httpHost = fmt.Sprintf("http://%s", os.Getenv(httpHostEnvVar))
)

type GRPCSuite struct {
	*testing.T
	URLClient urlshortenerv1.URLShortenerClient
}

func NewGRPCSuite(t *testing.T) (context.Context, *GRPCSuite) {
	t.Helper()

	const operationsTimeout = time.Second * 10

	ctx, cancel := context.WithTimeout(context.Background(), operationsTimeout)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	cc, err := grpc.NewClient(
		grpcHost,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("gRPC server connection failed: %v", err)
	}

	return ctx, &GRPCSuite{
		T:         t,
		URLClient: urlshortenerv1.NewURLShortenerClient(cc),
	}
}

type HTTPSuite struct {
	*testing.T
	Client  *http.Client
	BaseURL string
}

func NewHTTPSuite(t *testing.T) (context.Context, *HTTPSuite) {
	t.Helper()

	const operationsTimeout = time.Second * 10

	ctx, cancel := context.WithTimeout(context.Background(), operationsTimeout)

	t.Cleanup(func() {
		t.Helper()
		cancel()
	})

	client := &http.Client{
		Timeout: operationsTimeout,
	}

	return ctx, &HTTPSuite{
		T:       t,
		Client:  client,
		BaseURL: httpHost,
	}
}
