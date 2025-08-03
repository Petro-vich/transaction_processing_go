package wallet

import (
	"fmt"
	"testing"

	"github.com/Petro-vich/transaction_processing_go/internal/models/transaction"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStorage struct {
	mock.Mock
}

func (_m *mockStorage) CreateWallet(address string, amount float64) error {
	args := _m.Called(address, amount)
	fmt.Println(args.Error(0))
	return args.Error(0)
}

func (_m *mockStorage) GetBalance(address string) (float64, error) {
	args := _m.Called(address)
	return args.Get(0).(float64), args.Error(1)
}

func (_m *mockStorage) SendMoney(from, to string, amount float64) error {
	args := _m.Called(from, to, amount)
	return args.Error(0)
}

func (_m *mockStorage) GetLast(count int) ([]transaction.Request, error) {
	args := _m.Called(count)
	return args.Get(0).([]transaction.Request), args.Error(1)
}

func TestWalletService_Initialize(t *testing.T) {
	t.Run("Successful init", func(t *testing.T) {
		store := &mockStorage{}
		service := NewService(store)

		store.On("CreateWallet", mock.AnythingOfType("string"), 100.0).Return(nil).Times(2)

		err := service.InitWall(2)

		assert.NoError(t, err)
	})

}
