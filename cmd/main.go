package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/Petro-vich/transaction_processing_go/internal/config"
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

	//fmt.Println(config.StoragePath)
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
