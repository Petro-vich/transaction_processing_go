package wallet

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"

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
	if count <= 0 {
		return fmt.Errorf("count can not be zero or negative")
	}

	var wg sync.WaitGroup
	var chErr = make(chan error, count)
	var mu sync.Mutex

	for i := 0; count > i; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wallAdr, err := generateWalletAddress()
			if err != nil {
				chErr <- err
				return
			}
			mu.Lock()
			err = ws.storage.CreateWallet(wallAdr, 100.0)
			mu.Unlock()
			if err != nil {
				chErr <- err
				return
			}

		}()
	}

	wg.Wait()

	close(chErr)
	for err := range chErr {
		if err != nil {
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
