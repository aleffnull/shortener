package main

import (
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
	logger := logger.NewZapLogger(zap)

	configuration, err := config.GetConfiguration()
	if err != nil {
		logger.Fatalf("configuration error: %v", err)
	}

	logger.Infof("using configuration: %+v", configuration)

	storage := store.NewMemoryStore(configuration)
	shortener := app.NewShortenerApp(storage, configuration)
	handler := app.NewHandler(shortener)
	router := app.NewRouter(logger)
	router.Prepare(handler)

	err = router.Run(configuration)
	if err != nil {
		logger.Fatalf("failed to start server: %v", err)
	}
}
