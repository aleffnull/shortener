package app

import (
	"net/http"

	"github.com/aleffnull/shortener/internal/middleware"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	cm "github.com/go-chi/chi/v5/middleware"
	"github.com/ldez/mimetype"
)

const mimetypeApplicationGZIP = "application/x-gzip"

type Router struct {
	handler              *Handler
	authorizationService service.AuthorizationService
	logger               logger.Logger
}

func NewRouter(handler *Handler, authorizationService service.AuthorizationService, logger logger.Logger) *Router {
	return &Router{
		handler:              handler,
		authorizationService: authorizationService,
		logger:               logger,
	}
}

func (r *Router) NewMuxHandler() http.Handler {
	mux := chi.NewRouter()

	mux.Get("/ping",
		middleware.LogHandler(
			r.handler.HandlePingRequest,
			r.logger))

	mux.Get("/{key}",
		middleware.LogHandler(
			middleware.UserIDHandler(
				setContentType(
					func(writer http.ResponseWriter, request *http.Request) {
						key := chi.URLParam(request, "key")
						r.handler.HandleGetRequest(writer, request, key)
					},
					mimetype.TextPlain),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsRequireValidToken),
			r.logger))

	mux.Get("/api/user/urls",
		middleware.LogHandler(
			middleware.UserIDHandler(
				setContentType(
					middleware.GzipHandler(r.handler.HandleGetUserURLsRequest),
					mimetype.ApplicationJSON, mimetypeApplicationGZIP),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsRequireValidToken),
			r.logger))

	mux.Delete("/api/user/urls",
		middleware.LogHandler(
			middleware.UserIDHandler(
				setContentType(
					middleware.GzipHandler(r.handler.HandleBatchDeleteRequest),
					mimetype.ApplicationJSON, mimetypeApplicationGZIP),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsRequireValidToken),
			r.logger))

	mux.Post("/",
		middleware.LogHandler(
			middleware.UserIDHandler(
				setContentType(
					middleware.GzipHandler(r.handler.HandlePostRequest),
					mimetype.TextPlain, mimetypeApplicationGZIP),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsNone),
			r.logger))

	mux.Post("/api/shorten",
		middleware.LogHandler(
			middleware.UserIDHandler(
				setContentType(
					middleware.GzipHandler(r.handler.HandleAPIRequest),
					mimetype.ApplicationJSON, mimetypeApplicationGZIP),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsNone),
			r.logger))

	mux.Post("/api/shorten/batch",
		middleware.LogHandler(
			middleware.UserIDHandler(
				setContentType(
					middleware.GzipHandler(r.handler.HandleAPIBatchRequest),
					mimetype.ApplicationJSON, mimetypeApplicationGZIP),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsNone),
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
