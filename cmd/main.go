package main

import (
	"log/slog"
	"os"

	"github.com/Petro-vich/transaction_processing_go/internal/config"
	"github.com/Petro-vich/transaction_processing_go/internal/storage/sqlite"
)

const (
	envLocal = "local"
	envProd  = "prod"
	env      = "debug"
)

func main() {

	config := config.Load()
	log := setupSlog(config.Env)
	log.Info("Start of the program")
	log.Debug("Slog initialized")

	storage, err := sqlite.New(config.StoragePath)
	if err != nil {
		log.Error("failed to init storage")
	}
	_ = storage
}

func setupSlog(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
