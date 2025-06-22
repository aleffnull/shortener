package app

import (
	"net"
	"net/http"

	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/middleware"
	"github.com/go-chi/chi/v5"
	cm "github.com/go-chi/chi/v5/middleware"
	"github.com/ldez/mimetype"
	"golang.org/x/net/context"
)

const mimetypeApplicationGZIP = "application/x-gzip"

type Router struct {
	mux *chi.Mux
}

func NewRouter() *Router {
	return &Router{
		mux: chi.NewRouter(),
	}
}

func (r *Router) Prepare(handler *Handler) {
	r.mux.Get("/{key}",
		middleware.Log(
			setContentType(
				func(writer http.ResponseWriter, request *http.Request) {
					key := chi.URLParam(request, "key")
					handler.HandleGetRequest(writer, key)
				},
				mimetype.TextPlain)))

	r.mux.Post("/",
		middleware.Log(
			setContentType(
				middleware.GzipHandler(handler.HandlePostRequest),
				mimetype.TextPlain, mimetypeApplicationGZIP)))

	r.mux.Post("/api/shorten",
		middleware.Log(
			setContentType(
				middleware.GzipHandler(handler.HandleAPIRequest),
				mimetype.ApplicationJSON, mimetypeApplicationGZIP)))
}

func setContentType(next http.HandlerFunc, contentType ...string) http.HandlerFunc {
	restrictor := cm.AllowContentType(contentType...)
	restricted := restrictor(next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restricted.ServeHTTP(w, r)
	})
}

func (r *Router) Run(ctx context.Context, configuration *config.Configuration) error {
	server := &http.Server{
		Addr:    configuration.ServerAddress,
		Handler: r.mux,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	return server.ListenAndServe()
}
