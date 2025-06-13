package app

import (
	"net/http"
	"time"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/middleware"
	"github.com/go-chi/chi/v5"
	cm "github.com/go-chi/chi/v5/middleware"
	"github.com/ldez/mimetype"
)

type Router struct {
	mux    *chi.Mux
	logger logger.Logger
}

func NewRouter(logger logger.Logger) *Router {
	return &Router{
		mux:    chi.NewRouter(),
		logger: logger,
	}
}

func (r *Router) Prepare(handler *Handler) {
	r.mux.Use(cm.AllowContentType(mimetype.TextPlain))
	r.mux.Get("/{key}", r.loggingMiddleware(r.makeGetHandler(handler)))
	r.mux.Post("/", r.loggingMiddleware(r.makePostHandler(handler)))
}

func (r *Router) Run(configuration *config.Configuration) error {
	return http.ListenAndServe(configuration.ServerAddress, r.mux)
}

func (r *Router) makeGetHandler(handler *Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")
		handler.HandleGetRequest(writer, key)
	}
}

func (r *Router) makePostHandler(handler *Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		handler.HandlePostRequest(writer, request)
	}
}

func (r *Router) loggingMiddleware(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		startTime := time.Now()
		responseWriter := middleware.NewResponseWriter(writer)
		handlerFunc(responseWriter, request)
		duration := time.Since(startTime)

		r.logger.Infof(
			"URL: %v, method: %v, time: %v, status: %v, size %v bytes",
			request.URL,
			request.Method,
			duration,
			responseWriter.GetStatus(),
			responseWriter.GetSize())
	}
}
