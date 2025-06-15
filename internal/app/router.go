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
	setPlaintText := createContentTypeRestriction(mimetype.TextPlain)
	setJSON := createContentTypeRestriction(mimetype.ApplicationJSON)

	r.mux.Get("/{key}", r.doLog(setPlaintText(makeGetHandler(handler))))
	r.mux.Post("/", r.doLog(setPlaintText(handler.HandlePostRequest)))
	r.mux.Post("/api/shorten", r.doLog(setJSON(handler.HandleAPIRequest)))
}

func createContentTypeRestriction(contentType string) func(http.HandlerFunc) http.HandlerFunc {
	restrictor := cm.AllowContentType(contentType)
	return func(handler http.HandlerFunc) http.HandlerFunc {
		restricted := restrictor(handler)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			restricted.ServeHTTP(w, r)
		})
	}
}

func (r *Router) Run(configuration *config.Configuration) error {
	return http.ListenAndServe(configuration.ServerAddress, r.mux)
}

func (r *Router) doLog(handlerFunc http.HandlerFunc) http.HandlerFunc {
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

func makeGetHandler(handler *Handler) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		key := chi.URLParam(request, "key")
		handler.HandleGetRequest(writer, key)
	}
}
