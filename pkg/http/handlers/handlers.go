package handlers

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	httpSwagger "github.com/swaggo/http-swagger"
	"log/slog"
	"net/http"
	pkgmiddleware "ozon_task/pkg/http/middleware"
	"ozon_task/pkg/http/responses"
)

type Handler func(*http.Request) responses.Response

func NewHandler(basePath string, opts ...RouterOption) http.Handler {
	baseRouter := chi.NewRouter()
	baseRouter.Route(basePath, func(r chi.Router) {
		for _, opt := range opts {
			opt(r)
		}
	})
	return baseRouter
}

func AddHandler(
	mountMethod func(pattern string, h http.HandlerFunc),
	pattern string,
	handler Handler,
) {
	mountMethod(pattern, Converter(handler))
}

func Converter(h Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp := h(r)
		if resp == nil {
			return
		}

		writeResponse(w, r, resp)
	}
}

func writeResponse(w http.ResponseWriter, r *http.Request, response responses.Response) {
	render.Status(r, response.StatusCode())
	render.JSON(w, r, response.GetPayload())
}

func DecodeRequest(r *http.Request, v interface{}) error {
	return render.Decode(r, v)
}

type RouterOption func(chi.Router)

func RouterOptions(options ...RouterOption) func(chi.Router) {
	return func(r chi.Router) {
		for _, option := range options {
			option(r)
		}
	}
}

func WithHealthHandler() RouterOption {
	return func(r chi.Router) {
		r.Mount("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			render.Status(r, http.StatusOK)
			render.PlainText(w, r, "OK")
		}))
	}
}

func WithSwagger() RouterOption {
	return func(r chi.Router) {
		r.Get("/docs/*", httpSwagger.Handler(
			httpSwagger.URL("docs/doc.json"),
		))
	}
}

func WithErrHandlers() RouterOption {
	return func(r chi.Router) {
		r.NotFound(Converter(
			func(r *http.Request) responses.Response {
				return responses.NotFound(errors.New("no such path"))
			}))
		r.MethodNotAllowed(Converter(
			func(r *http.Request) responses.Response {
				return responses.MethodNotAllowed(errors.New("this method is not allowed"))
			}))
	}
}

func WithLogging(logger *slog.Logger) RouterOption {
	return func(r chi.Router) {
		r.Use(pkgmiddleware.NewLoggingMiddleware(logger))
	}
}

func WithRecover() RouterOption {
	return func(r chi.Router) {
		r.Use(middleware.Recoverer)
	}
}

func WithProfilerHandlers() RouterOption {
	return func(r chi.Router) {
		middleware.Profiler()
	}
}

func WithRequestID() RouterOption {
	return func(r chi.Router) {
		r.Use(middleware.RequestID)
	}
}

func WithMetricsHandler() RouterOption {
	return func(r chi.Router) {
		r.Mount("/metrics", promhttp.Handler())
	}
}
