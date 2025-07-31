package main

import (
	"log/slog"
	"os"

	"github.com/Petro-vich/transaction_processing_go/internal/config"
	httpserver "github.com/Petro-vich/transaction_processing_go/internal/http-server"
	"github.com/Petro-vich/transaction_processing_go/internal/lib/logger/sl"
	"github.com/Petro-vich/transaction_processing_go/internal/storage/sqlite"
)

func main() {

	config := config.Load()
	log := sl.SetupSlog(config.Env)
	log.Info("Start of the program")
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New("/home/gaylesga/Desktop/s21.project/go/transaction_processing_go/storage/sqlite/storage.db")
	//config.StoragePath
	//TODO:
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	server := httpserver.New(storage, config, log)
	log.Info("Starting server:", slog.String("address", config.Address))
	if err := server.Start(); err != nil {
		log.Error("failes to start server", sl.Err(err))
		os.Exit(1)
	}

	// walletInit := wallet.NewInitializer(storage)
	// if err := walletInit.Initializer(10); err != nil { //TODO: yaml: `start_wallet_poop`
	// 	log.Error("failed to init pool wallets", sl.Err(err))
	// }
	// log.Info("the starter set of wallets has been added")

	// adr := "70200aebb271f81246dc200bbf8da3670f1568d541279cb32b822d454985777"
	// balance, err := storage.GetBalance(adr)
	// if err != nil {
	// 	log.Error("failed to get balance", sl.Err(err))
	// }
	// fmt.Println(balance)

}
