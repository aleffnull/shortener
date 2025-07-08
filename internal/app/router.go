package app

import (
	"net/http"

	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/middleware"
	"github.com/go-chi/chi/v5"
	cm "github.com/go-chi/chi/v5/middleware"
	"github.com/ldez/mimetype"
)

const mimetypeApplicationGZIP = "application/x-gzip"

type Router struct {
	handler *Handler
	logger  logger.Logger
}

func NewRouter(handler *Handler, logger logger.Logger) *Router {
	return &Router{
		handler: handler,
		logger:  logger,
	}
}

func (r *Router) NewMuxHandler() http.Handler {
	mux := chi.NewRouter()

	mux.Get("/ping",
		middleware.Log(
			r.handler.HandlePingRequest,
			r.logger))

	mux.Get("/{key}",
		middleware.Log(
			setContentType(
				func(writer http.ResponseWriter, request *http.Request) {
					key := chi.URLParam(request, "key")
					r.handler.HandleGetRequest(writer, key)
				},
				mimetype.TextPlain),
			r.logger))

	mux.Post("/",
		middleware.Log(
			setContentType(
				middleware.GzipHandler(r.handler.HandlePostRequest),
				mimetype.TextPlain, mimetypeApplicationGZIP),
			r.logger))

	mux.Post("/api/shorten",
		middleware.Log(
			setContentType(
				middleware.GzipHandler(r.handler.HandleAPIRequest),
				mimetype.ApplicationJSON, mimetypeApplicationGZIP),
			r.logger))

	return mux
}

func setContentType(next http.HandlerFunc, contentType ...string) http.HandlerFunc {
	restrictor := cm.AllowContentType(contentType...)
	restricted := restrictor(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restricted.ServeHTTP(w, r)
	})
}
