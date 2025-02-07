package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"
)

const shutdownTimeout = time.Second * 10

func NewServer(addr string, handler http.Handler, readTimeout, writeTimeout, idleTimeout time.Duration) *http.Server {
	return &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

func RunServer(ctx context.Context, addr string, handler http.Handler, readTimeout, writeTimeout, idleTimeout time.Duration) error {
	const op = "server.RunServer"

	slog.With(slog.String("op", op), slog.String("address", addr)).Info("Starting http server")
	server := NewServer(addr, handler, readTimeout, writeTimeout, idleTimeout)

	errListen := make(chan error, 1)
	go func() {
		errListen <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err := server.Shutdown(ctxShutdown)
		if err != nil {
			return fmt.Errorf("can't shutdown server: %w", err)
		}
		return nil
	case err := <-errListen:
		return fmt.Errorf("can't run server: %w", err)
	}
}
