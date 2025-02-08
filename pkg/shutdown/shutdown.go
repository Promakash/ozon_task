package shutdown

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

var ErrOSSignal = errors.New("operating system signal")

func ListenSignal(ctx context.Context, logger *slog.Logger) error {
	sigquit := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(sigquit, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-ctx.Done():
		return nil
	case sig := <-sigquit:
		logger.Info("Captured signal: ", slog.String("signal", sig.String()))
		return ErrOSSignal
	}
}
