package sqlite

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage_CreateWallet(t *testing.T) {
	store, err := New(":memory:")
	assert.NoError(t, err)
	defer store.db.Close()

	t.Run("Valid wallet", func(t *testing.T) {
		addr := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		err := store.CreateWallet(addr, 100.0)
		assert.NoError(t, err)

		balance, err := store.GetBalance(addr)
		assert.NoError(t, err)
		assert.Equal(t, 100.0, balance)
	})

	t.Run("Invalid address length", func(t *testing.T) {
		err := store.CreateWallet("short", 100.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid address length")
	})

	t.Run("Negative balance", func(t *testing.T) {
		addr := "2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		err := store.CreateWallet(addr, -100.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "balanc must be positiv")
	})

	t.Run("Duplicate address", func(t *testing.T) {
		addr := "3234567890abcdef1234567890accdef1234567890abcdef1234567890abcdef"
		err := store.CreateWallet(addr, 100.0)
		assert.NoError(t, err)

		err = store.CreateWallet(addr, 200.0)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "UNIQUE constraint failed")
	})
}

func TestStorage_SendMoney(t *testing.T) {
	store, err := New(":memory:")
	assert.NoError(t, err)
	defer store.db.Close()

	fromAddr := "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
	toAddr := "2234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

	err = store.CreateWallet(fromAddr, 200.0)
	assert.NoError(t, err)
	err = store.CreateWallet(toAddr, 50.0)
	assert.NoError(t, err)

	t.Run("Successful send money", func(t *testing.T) {
		err := store.SendMoney(fromAddr, toAddr, 100.0)
		assert.NoError(t, err)

		fromBalance, err := store.GetBalance(fromAddr)
		assert.NoError(t, err)
		assert.Equal(t, 100.0, fromBalance, "Sender balance should be 100.0 after sending 100.0")

		toBalance, err := store.GetBalance(toAddr)
		assert.NoError(t, err)
		assert.Equal(t, 150.0, toBalance, "Receiver balance should be 150.0 after receiving 100.0")
	})
}

// 	t.Run("Insufficient funds", func(t *testing.T) {
// 		err := store.SendMoney(fromAddr, toAddr, 200.0)
// 		assert.ErrorIs(t, err, storage.ErrInsufficient)

// 		fromBalance, err := store.GetBalance(fromAddr)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 100.0, fromBalance, "Sender balance should remain unchanged")

// 		toBalance, err := store.GetBalance(toAddr)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 150.0, toBalance, "Receiver balance should remain unchanged")
// 	})

// 	t.Run("Non-existent from address", func(t *testing.T) {
// 		nonExistentAddr := "3234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
// 		err := store.SendMoney(nonExistentAddr, toAddr, 50.0)
// 		assert.ErrorIs(t, err, storage.ErrAddressNotExist)
// 	})

// 	t.Run("Non-existent to address", func(t *testing.T) {
// 		nonExistentAddr := "3234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
// 		err := store.SendMoney(fromAddr, nonExistentAddr, 50.0)
// 		assert.ErrorIs(t, err, storage.ErrAddressNotExist)
// 	})

// 	t.Run("Invalid from address length", func(t *testing.T) {
// 		err := store.SendMoney("short", toAddr, 50.0)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "invalid address length")
// 	})

// 	t.Run("Invalid to address length", func(t *testing.T) {
// 		err := store.SendMoney(fromAddr, "short", 50.0)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "invalid address length")
// 	})

// 	t.Run("Non-positive amount", func(t *testing.T) {
// 		err := store.SendMoney(fromAddr, toAddr, 0)
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "amount must be positive")
// 	})
