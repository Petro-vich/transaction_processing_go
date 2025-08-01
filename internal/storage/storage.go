package storage

import (
	"errors"

	"github.com/Petro-vich/transaction_processing_go/internal/models/transaction"
)

var (
	ErrAddressNotExist = errors.New("the address does not exist")
	ErrInsufficient    = errors.New("insufficient funds")
)

type Repository interface {
	CreateWallet(address string, amount float64) error
	GetBalance(address string) (float64, error)
	SendMoney(from, to string, amount float64) error
	GetLast(count int) ([]transaction.Request, error)
}
