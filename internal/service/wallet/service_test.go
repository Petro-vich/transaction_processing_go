package wallet

import (
	"testing"

	"github.com/Petro-vich/transaction_processing_go/internal/models/transaction"
	"github.com/Petro-vich/transaction_processing_go/internal/storage/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStorage struct {
	mock.Mock
}

func (_m *mockStorage) CreateWallet(address string, amount float64) error {
	args := _m.Called(address, amount)
	//fmt.Println(args.Error(0))
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
	store := &mockStorage{}
	service := NewService(store)
	t.Run("Successful init", func(t *testing.T) {

		store.On("CreateWallet", mock.AnythingOfType("string"), 100.0).Return(nil).Times(1)

		err := service.InitWall(1)
		assert.NoError(t, err)
	})

	t.Run("Negative count", func(t *testing.T) {

		err := service.InitWall(-1)
		assert.EqualError(t, err, "count can not be zero or negative")
	})

	t.Run("zero count", func(t *testing.T) {

		err := service.InitWall(-1)
		assert.EqualError(t, err, "count can not be zero or negative")
	})
}

func BenchmarkInitWallSequential(b *testing.B) {
	store, err := sqlite.New("file::memory:?cache=shared")
	if err != nil {
		b.Fatalf("failed to create test DB: %v", err)
	}

	service := NewService(store)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.InitWall(1000) // Тестируем с 1000 кошельков
	}
}
