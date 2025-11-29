package app

import (
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	cm "github.com/go-chi/chi/v5/middleware"
	"github.com/ldez/mimetype"

	"github.com/aleffnull/shortener/internal/middleware"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/service"
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
	mux.Use(func(next http.Handler) http.Handler {
		return middleware.LogHandler(next, r.logger)
	})

	mux.Get("/ping", r.handler.HandlePingRequest)

	mux.Route("/", func(t chi.Router) {
		t.Get("/{key}",
			middleware.UserIDHandler(
				setContentType(
					func(writer http.ResponseWriter, request *http.Request) {
						key := chi.URLParam(request, "key")
						r.handler.HandleGetRequest(writer, request, key)
					},
					mimetype.TextPlain),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsRequireValidToken))
		t.Post("/",
			middleware.UserIDHandler(
				setContentType(
					middleware.GzipHandler(r.handler.HandlePostRequest),
					mimetype.TextPlain, mimetypeApplicationGZIP),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsNone))
	})

	mux.Route("/api/shorten", func(t chi.Router) {
		t.Use(func(next http.Handler) http.Handler {
			return middleware.UserIDHandler(
				setContentType(
					middleware.GzipHandler(next.ServeHTTP),
					mimetype.ApplicationJSON, mimetypeApplicationGZIP),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsNone)
		})
		t.Post("/", r.handler.HandleAPIRequest)
		t.Post("/batch", r.handler.HandleAPIBatchRequest)
	})

	mux.Route("/api/user/urls", func(t chi.Router) {
		t.Use(func(next http.Handler) http.Handler {
			return middleware.UserIDHandler(
				setContentType(
					middleware.GzipHandler(next.ServeHTTP),
					mimetype.ApplicationJSON, mimetypeApplicationGZIP),
				r.authorizationService,
				r.logger,
				middleware.UserIDOptionsRequireValidToken)
		})

		t.Get("/", r.handler.HandleGetUserURLsRequest)
		t.Delete("/", r.handler.HandleBatchDeleteRequest)
	})

	mux.Route("/debug/pprof", func(r chi.Router) {
		r.Get("/cmdline", pprof.Cmdline)
		r.Get("/profile", pprof.Profile)
		r.Get("/symbol", pprof.Symbol)
		r.Get("/trace", pprof.Trace)
		r.Get("/*", pprof.Index)
	})

	return mux
}

func setContentType(next http.HandlerFunc, contentType ...string) http.HandlerFunc {
	restrictor := cm.AllowContentType(contentType...)
	restricted := restrictor(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restricted.ServeHTTP(w, r)
	})
}
