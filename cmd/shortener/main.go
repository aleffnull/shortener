package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"runtime/pprof"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"

	"github.com/aleffnull/shortener/internal/app"
	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/audit"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/parameters"
	"github.com/aleffnull/shortener/internal/pkg/store"
	"github.com/aleffnull/shortener/internal/repository"
	"github.com/aleffnull/shortener/internal/service"
)

func NewShortenerApp(
	lc fx.Lifecycle,
	connection repository.Connection,
	storage store.Store,
	deleteURLsService service.DeleteURLsService,
	auditService service.AuditService,
	log logger.Logger,
	parameters parameters.AppParameters,
	configuration *config.Configuration,
) app.App {
	shortener := app.NewShortenerApp(connection, storage, deleteURLsService, auditService, log, parameters, configuration)
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
	var cpuProfile *os.File
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if cpu, err := startCPUProfile(configuration.CPUProfile); err != nil {
				return err
			} else {
				cpuProfile = cpu
			}

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
			if cpuProfile != nil {
				if err := stopCPUProfile(cpuProfile); err != nil {
					log.Errorf("failed to stop CPU profiling: %w", err)
				}
			}

			if err := collectMemoryProfile(configuration.MemoryProfile); err != nil {
				log.Errorf("failed to collect memory profile: %w", err)
			}

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
			service.NewDeleteURLsService,
			fx.Annotate(service.NewAuditService, fx.ParamTags(`group:"receivers"`)),
			NewShortenerApp,
			app.NewRouter,
			NewHTTPServer,
			app.NewMaintenanceHandler,
			app.NewSimpleAPIHandler,
			app.NewAPIHandler,
			app.NewUserHandler,
			asReceiver(audit.NewFileReceiver),
			asReceiver(audit.NewEndpointReceiver),
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

func asReceiver(f any) any {
	return fx.Annotate(f, fx.As(new(audit.Receiver)), fx.ResultTags(`group:"receivers"`))
}

func startCPUProfile(filePath string) (*os.File, error) {
	if len(filePath) == 0 {
		return nil, nil
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create CPU profile file: %w", err)
	}

	err = pprof.StartCPUProfile(file)
	if err != nil {
		err = errors.Join(err, file.Close())
		return nil, fmt.Errorf("failed to start CPU profiling: %w", err)
	}

	return file, nil
}

func stopCPUProfile(file *os.File) error {
	pprof.StopCPUProfile()
	return file.Close()
}

func collectMemoryProfile(filePath string) (err error) {
	if len(filePath) == 0 {
		return nil
	}

	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to create memory profile file: %w", err)
	}

	defer func() {
		err = errors.Join(err, file.Close())
	}()

	runtime.GC()
	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("failed to write memory profile to file: %w", err)
	}

	return nil
}
