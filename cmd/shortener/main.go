package main

import (
	"context"
	"log"

	"github.com/aleffnull/shortener/internal/app"
	"github.com/aleffnull/shortener/internal/config"
	"github.com/aleffnull/shortener/internal/pkg/logger"
	"github.com/aleffnull/shortener/internal/store"
	"go.uber.org/zap"
)

func main() {
	zap, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("failed to create zap logger: %v", err)
	}

	defer zap.Sync()
	log := logger.NewZapLogger(zap)

	configuration, err := config.GetConfiguration()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	log.Infof("using configuration: %+v", configuration)

	ctx := context.Background()
	ctx = logger.ContextWithLogger(ctx, log)

	storage := store.NewMemoryStore(configuration)
	coldStorage := store.NewFileStore(configuration)
	shortener := app.NewShortenerApp(storage, coldStorage, configuration)
	if err = shortener.Init(ctx); err != nil {
		log.Fatalf("failed to init application: %v", err)
	}

	handler := app.NewHandler(shortener)
	router := app.NewRouter()
	router.Prepare(handler)

	err = router.Run(ctx, configuration)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
