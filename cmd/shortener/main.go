package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"syscall"

	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/aleffnull/shortener/internal/app"
	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/middleware"
	"github.com/aleffnull/shortener/internal/pkg/audit"
	app_grpc "github.com/aleffnull/shortener/internal/pkg/grpc"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/pkg/parameters"
	"github.com/aleffnull/shortener/internal/pkg/pb/shortener/api"
	"github.com/aleffnull/shortener/internal/pkg/store"
	"github.com/aleffnull/shortener/internal/repository"
	"github.com/aleffnull/shortener/internal/service"
	"github.com/samber/lo"
)

var (
	BuildVersion string
	BuildDate    string
	BuildCommit  string
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
			log.Infof("Start application")
			err := shortener.Init(ctx)
			if err != nil {
				return fmt.Errorf("application initialization failed: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Infof("Stop application")
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
	stdLogProvider logger.StdLogProvider,
) *http.Server {
	srv := &http.Server{
		Addr:     configuration.ServerAddress,
		Handler:  router.NewMuxHandler(),
		ErrorLog: stdLogProvider.GetStdLog(),
	}
	var cpuProfile *os.File
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if cpu, err := startCPUProfile(configuration.CPUProfile); err != nil {
				return err
			} else {
				cpuProfile = cpu
			}

			log.Infof("Starting HTTP%v server on %v", lo.Ternary(configuration.HTTPS.Enabled, "S", ""), srv.Addr)
			go func() {
				err := lo.Ternary(
					configuration.HTTPS.Enabled,
					srv.ListenAndServeTLS(configuration.HTTPS.CertificateFile, configuration.HTTPS.KeyFile),
					srv.ListenAndServe(),
				)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatalf("Server start error: %v", err)
				}
			}()

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

			log.Infof("Stop server")
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func NewGRPCServer(
	lc fx.Lifecycle,
	service *app_grpc.ShortenerService,
	authorizationService service.AuthorizationService,
	configuration *config.Configuration,
	log logger.Logger,
) *grpc.Server {
	var server *grpc.Server
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			log.Infof("Starting GRPC server on %v", configuration.ServerAddressGRPC)
			go func() {
				// GRPC сервер не является основным, поэтому не роняем здесь сервис при ошибках.

				listener, err := net.Listen("tcp", configuration.ServerAddressGRPC)
				if err != nil {
					log.Warnf("Failed to start listening on GRPC service port: %v", err)
					return
				}

				interceptor := func(
					ctx context.Context,
					request any,
					info *grpc.UnaryServerInfo,
					handler grpc.UnaryHandler,
				) (any, error) {
					return middleware.UserIDInterceptor(ctx, request, info, handler, authorizationService, log)
				}
				server = grpc.NewServer(grpc.UnaryInterceptor(interceptor))
				api.RegisterShortenerServiceServer(server, service)
				if err = server.Serve(listener); err != nil {
					log.Warnf("GRPC server error: %w", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			if server != nil {
				log.Infof("Stop GRPC server")
				server.Stop()
			}

			return nil
		},
	})

	return server
}

func main() {
	fmt.Printf("Build version: %v\n", getValueOrNA(BuildVersion))
	fmt.Printf("Build date: %v\n", getValueOrNA(BuildDate))
	fmt.Printf("Build commit: %v\n", getValueOrNA(BuildCommit))

	fx.New(
		fx.Provide(
			zap.NewDevelopment,
			logger.NewZapLogger,
			logger.NewStdLogProvider,
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
			app.NewInternalHandler,
			app_grpc.NewShortenerService,
			NewGRPCServer,
			asReceiver(audit.NewFileReceiver),
			asReceiver(audit.NewEndpointReceiver),
		),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log}
		}),
		fx.Invoke(func(
			_ *http.Server,
			_ *grpc.Server,
			configuration *config.Configuration,
			logger logger.Logger,
			shutdowner fx.Shutdowner,
		) {
			logger.Infof("Using configuration: %v", configuration)
			connectToSignals(logger, shutdowner)
		}),
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

func getValueOrNA(value string) string {
	return lo.Ternary(len(value) == 0, "N/A", value)
}

func connectToSignals(logger logger.Logger, shutdowner fx.Shutdowner) {
	// syscall.SIGTERM и syscall.SIGINT обрабатываются контейнером Uber FX.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGQUIT)
	go func() {
		for {
			<-ctx.Done()
			logger.Infof("Got SIGQUIT signal")
			cancel()
			shutdowner.Shutdown()
			break
		}
	}()
}
