package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/aleffnull/shortener/internal/app"
	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/parameters"
	"github.com/aleffnull/shortener/internal/pkg/store"
	"github.com/aleffnull/shortener/internal/repository"
	"github.com/aleffnull/shortener/internal/service"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func NewShortenerApp(
	lc fx.Lifecycle,
	connection repository.Connection,
	storage store.Store,
	log logger.Logger,
	parameters parameters.AppParameters,
	configuration *config.Configuration,
) app.App {
	shortener := app.NewShortenerApp(connection, storage, log, parameters, configuration)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			err := shortener.Init(ctx)
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
	fx.New(
		fx.Provide(
			zap.NewDevelopment,
			logger.NewZapLogger,
			config.GetConfiguration,
			repository.NewConnection,
			store.NewFileStore,
			store.NewStore,
			parameters.NewAppParameters,
			service.NewAuthorizationService,
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
