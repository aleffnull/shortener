package main

import (
	"context"
	"net"
	"net/http"

	"github.com/aleffnull/shortener/internal/app"
	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/store"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewShortenerApp(
	lc fx.Lifecycle,
	storage store.Store,
	coldStorage store.ColdStore,
	log logger.Logger,
	configuration *config.Configuration,
) app.App {
	shortener := app.NewShortenerApp(storage, coldStorage, log, configuration)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			shortener.Init(ctx)
			return nil
		},
	})

	return shortener
}

func NewHTTPServer(
	lc fx.Lifecycle,
	router *app.Router,
	configuration *config.Configuration,
	log *zap.Logger,
) *http.Server {
	srv := &http.Server{
		Addr:    configuration.ServerAddress,
		Handler: router.NewMuxHandler(),
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			listener, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}

			log.Info("Starting HTTP server", zap.String("addr", srv.Addr))
			go srv.Serve(listener)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func main() {
	fx.New(
		fx.Provide(
			zap.NewDevelopment,
			logger.NewZapLogger,
			config.GetConfiguration,
			store.NewMemoryStore,
			store.NewFileStore,
			NewShortenerApp,
			app.NewHandler,
			app.NewRouter,
			NewHTTPServer,
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}
