package wallet

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/Petro-vich/transaction_processing_go/internal/storage"
)

type WalletService struct {
	storage storage.Repository
}

func NewService(storage storage.Repository) *WalletService {
	return &WalletService{
		storage: storage}
}

func (ws *WalletService) InitWall(count int) error {
	for i := 0; count > i; i++ {
		wallAdr, err := generateWalletAddress()
		if err != nil {
			return err
		}
		if err := ws.storage.CreateWallet(wallAdr, 100.0); err != nil {
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
