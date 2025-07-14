package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/aleffnull/shortener/internal/app"
	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/store"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewShortenerApp(
	lc fx.Lifecycle,
	storage store.Store,
	log logger.Logger,
	configuration *config.Configuration,
) app.App {
	shortener := app.NewShortenerApp(storage, log, configuration)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := shortener.Init()
			if err != nil {
				return fmt.Errorf("application initialization failed: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			shortener.Shutdown()
			return nil
		},
	})

	return shortener
}

func NewHTTPServer(
	lc fx.Lifecycle,
	router *app.Router,
	configuration *config.Configuration,
	log logger.Logger,
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

			log.Infof("Using configuration: %v", configuration)
			log.Infof("Starting HTTP server on %v", srv.Addr)
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
	configuration, err := config.GetConfiguration()
	if err != nil {
		log.Fatalf("Failed to get application configuration: %s", err)
	}

	configurationProvider := func() *config.Configuration { return configuration }
	storeProvider := func(coldStore store.ColdStore, logger logger.Logger) store.Store {
		if configuration.DatabaseStore.IsDatabaseEnabled() {
			return store.NewDatabaseStore(configuration, logger)
		}

		return store.NewMemoryStore(coldStore, configuration, logger)
	}

	fx.New(
		fx.Provide(
			zap.NewDevelopment,
			logger.NewZapLogger,
			configurationProvider,
			storeProvider,
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
