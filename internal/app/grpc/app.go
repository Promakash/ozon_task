package grpc

import (
	"fmt"
	"log/slog"
	"net"
	"ozon_task/internal/config"
	"ozon_task/internal/grpc/url_shortener"
	"ozon_task/internal/usecases"
	pkggrpc "ozon_task/pkg/grpc"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(
	log *slog.Logger,
	service usecases.URL,
	cfg config.GRPCConfig,
) *App {
	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.PayloadReceived, logging.PayloadSent,
		),
	}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandler(func(p interface{}) error {
			log.Error("Recovered from panic", slog.Any("panic", p))

			return status.Errorf(codes.Internal, "internal error")
		}),
	}

	gRPCServer := grpc.NewServer(grpc.ChainUnaryInterceptor(
		recovery.UnaryServerInterceptor(recoveryOpts...),
		logging.UnaryServerInterceptor(pkggrpc.InterceptorLogger(log), loggingOpts...),
	))

	url_shortener.Register(
		gRPCServer,
		service,
		cfg.OperationsTimeout,
		log,
	)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       cfg.Port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpc.App"

	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server has started", slog.String("addr", l.Addr().String()))

	if err = a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grcapp.Stop"

	log := a.log.With(slog.String("op", op))
	log.Info("gracefully stopping server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
	log.Info("gRPC server was stopped")
}
