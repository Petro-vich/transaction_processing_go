package wallet

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/Petro-vich/transaction_processing_go/internal/storage"
)

type Initialization struct {
	storage storage.Repository
}

func NewInitializer(storage storage.Repository) *Initialization {
	return &Initialization{storage: storage}
}

func (in *Initialization) Initializer(count int) error {
	for i := 0; count > i; i++ {
		wallAdr, err := generateWalletAddress()
		if err != nil {
			return err
		}
		if err := in.storage.CreateWallet(wallAdr, 100.0); err != nil {
			return err
		}
	}
	return nil
}

func generateWalletAddress() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
