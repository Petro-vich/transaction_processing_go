package sqlite

import (
	"strings"
	"testing"

	"github.com/Petro-vich/transaction_processing_go/internal/storage"
	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) *Storage {
	st, err := New("file::memory:?cache=shared")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}
	return st
}

func generateTestAddress(t *testing.T, prefix string) string {
	if len(prefix) > 64 {
		t.Fatalf("prefix too long for address: %s", prefix)
	}
	return prefix + strings.Repeat("0", 64-len(prefix))
}

func TestStorage_CreateWallet(t *testing.T) {
	t.Run("Successful wallet creation", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		address := generateTestAddress(t, "a")
		amount := 100.0

		err := st.CreateWallet(address, amount)
		assert.NoError(t, err)

		// Проверяем, что кошелек создан
		balance, err := st.GetBalance(address)
		assert.NoError(t, err)
		assert.Equal(t, amount, balance)
	})

	t.Run("Invalid address length", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		err := st.CreateWallet("short_address", 100.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid address length")
	})

	t.Run("Negative amount", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		err := st.CreateWallet(generateTestAddress(t, "a"), -10.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "balanc must be positiv")
	})
}

func TestStorage_GetBalance(t *testing.T) {
	t.Run("Get balance for existing wallet", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		address := generateTestAddress(t, "a")
		amount := 50.0

		err := st.CreateWallet(address, amount)
		assert.NoError(t, err)

		balance, err := st.GetBalance(address)
		assert.NoError(t, err)
		assert.Equal(t, amount, balance)
	})

	t.Run("Get balance for non-existent wallet", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		balance, err := st.GetBalance(generateTestAddress(t, "a"))
		assert.ErrorIs(t, err, storage.ErrAddressNotExist)
		assert.Equal(t, 0.0, balance)
	})
}

func TestStorage_SendMoney(t *testing.T) {
	t.Run("Successful transaction", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		fromAddr := generateTestAddress(t, "a")
		toAddr := generateTestAddress(t, "b")
		amount := 30.0

		err := st.CreateWallet(fromAddr, 100.0)
		assert.NoError(t, err)
		err = st.CreateWallet(toAddr, 50.0)
		assert.NoError(t, err)

		err = st.SendMoney(fromAddr, toAddr, amount)
		assert.NoError(t, err)

		fromBalance, err := st.GetBalance(fromAddr)
		assert.NoError(t, err)
		assert.Equal(t, 70.0, fromBalance)

		toBalance, err := st.GetBalance(toAddr)
		assert.NoError(t, err)
		assert.Equal(t, 80.0, toBalance)
	})

	t.Run("Insufficient funds", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		fromAddr := generateTestAddress(t, "a")
		toAddr := generateTestAddress(t, "b")

		err := st.CreateWallet(fromAddr, 20.0)
		assert.NoError(t, err)
		err = st.CreateWallet(toAddr, 50.0)
		assert.NoError(t, err)

		err = st.SendMoney(fromAddr, toAddr, 30.0)
		assert.ErrorIs(t, err, storage.ErrInsufficient)
	})

	t.Run("Non-existent from address", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		toAddr := generateTestAddress(t, "b")
		err := st.CreateWallet(toAddr, 50.0)
		assert.NoError(t, err)

		err = st.SendMoney(generateTestAddress(t, "a"), toAddr, 10.0)
		assert.ErrorIs(t, err, storage.ErrAddressNotExist)
	})
}

func TestStorage_GetLast(t *testing.T) {
	t.Run("Get last transactions", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		fromAddr := generateTestAddress(t, "a")
		toAddr := generateTestAddress(t, "b")
		amount := 30.0

		err := st.CreateWallet(fromAddr, 100.0)
		assert.NoError(t, err)
		err = st.CreateWallet(toAddr, 50.0)
		assert.NoError(t, err)

		err = st.SendMoney(fromAddr, toAddr, amount)
		assert.NoError(t, err)

		transactions, err := st.GetLast(1)
		assert.NoError(t, err)
		assert.Len(t, transactions, 1)
		assert.Equal(t, fromAddr, transactions[0].From)
		assert.Equal(t, toAddr, transactions[0].To)
		assert.Equal(t, amount, transactions[0].Amount)
	})

	t.Run("Invalid count", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		transactions, err := st.GetLast(0)
		assert.NoError(t, err)
		assert.Empty(t, transactions)
	})
}

func TestStorage_IsEmpty(t *testing.T) {
	t.Run("Empty storage", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		assert.True(t, st.IsEmpty())
	})

	t.Run("Non-empty storage", func(t *testing.T) {
		st := setupTestDB(t)
		defer st.db.Close()

		err := st.CreateWallet(generateTestAddress(t, "a"), 100.0)
		assert.NoError(t, err)

		assert.False(t, st.IsEmpty())
	})
}