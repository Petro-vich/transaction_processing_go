package main

import (
	"github.com/Petro-vich/transaction_processing_go/internal/config"
	"github.com/Petro-vich/transaction_processing_go/internal/lib/logger/sl"
	"github.com/Petro-vich/transaction_processing_go/internal/service/wallet"
	"github.com/Petro-vich/transaction_processing_go/internal/storage/sqlite"
)

func main() {
	config := config.Load()
	log := sl.SetupSlog(config.Env)
	log.Info("Start of the program")
	log.Debug("Slog initialized")

	storage, err := sqlite.New(config.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
	}

	walletInit := wallet.NewInitializer(storage)
	if err := walletInit.Initializer(10); err != nil {
		log.Error("failed to init pool wallets", sl.Err(err))
	}
	log.Info("the starter set of wallets has been added")
}
