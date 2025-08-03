package main

import (
	"log/slog"
	"os"

	"github.com/Petro-vich/transaction_processing_go/internal/config"
	httpserver "github.com/Petro-vich/transaction_processing_go/internal/http-server"
	"github.com/Petro-vich/transaction_processing_go/internal/lib/logger/sl"
	"github.com/Petro-vich/transaction_processing_go/internal/service/wallet"
	"github.com/Petro-vich/transaction_processing_go/internal/storage/sqlite"
)

func main() {

	cfg := config.Load()
	log := sl.SetupSlog(cfg.Env)
	log.Info("Start of the program")
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)

	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	if storage.IsEmpty() {
		WallServ := wallet.NewService(storage)
		if err := WallServ.InitWall(10); err != nil {
			log.Error("failed to init pool wallets", sl.Err(err))
		}
		log.Info("the starter set of wallets has been added")
	}

	server := httpserver.New(storage, cfg, log)
	log.Info("Starting server:", slog.String("address", cfg.Address))
	if err := server.Start(); err != nil {
		log.Error("failed to start server", sl.Err(err))

		os.Exit(1)
	}

}
