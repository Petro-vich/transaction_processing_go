package httpserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest" // Добавьте этот импорт
	"strings"
	"testing"

	"github.com/Petro-vich/transaction_processing_go/internal/config"
	"github.com/Petro-vich/transaction_processing_go/internal/lib/logger/sl"
	"github.com/Petro-vich/transaction_processing_go/internal/models/transaction"
	"github.com/Petro-vich/transaction_processing_go/internal/storage"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Повторно используем mockStorage из service_test.go
type mockStorage struct {
	mock.Mock
}

func (m *mockStorage) CreateWallet(address string, amount float64) error {
	args := m.Called(address, amount)
	return args.Error(0)
}

func (m *mockStorage) GetBalance(address string) (float64, error) {
	args := m.Called(address)
	return args.Get(0).(float64), args.Error(1)
}

func (m *mockStorage) SendMoney(from, to string, amount float64) error {
	args := m.Called(from, to, amount)
	return args.Error(0)
}

func (m *mockStorage) GetLast(count int) ([]transaction.Request, error) {
	args := m.Called(count)
	return args.Get(0).([]transaction.Request), args.Error(1)
}

// Вспомогательная функция для создания тестового сервера
func setupTestServer(t *testing.T, storage storage.Repository) *Server {
	cfg := &config.Config{
		Env: "test",
		HTTPServer: config.HTTPServer{
			Address: "localhost:8080",
		},
	}
	log := sl.SetupSlog("test")
	return New(storage, cfg, log)
}

// Вспомогательная функция для генерации адреса длиной 64 символа
func generateTestAddress(prefix string) string {
	return prefix + strings.Repeat("0", 64-len(prefix))
}

// Тесты для GetBalanceHandler
func TestGetBalanceHandler(t *testing.T) {
	t.Run("Successful balance retrieval", func(t *testing.T) {
		// Создаем мок хранилища
		store := &mockStorage{}
		server := setupTestServer(t, store)

		// Настраиваем мок
		address := generateTestAddress("a")
		store.On("GetBalance", address).Return(100.0, nil)

		// Создаем тестовый запрос
		req := httptest.NewRequest(http.MethodGet, "/api/wallet/"+address+"/balance", nil)
		req = mux.SetURLVars(req, map[string]string{"address": address})
		rr := httptest.NewRecorder()

		// Выполняем хендлер
		server.GetBalanceHandler(rr, req)

		// Проверяем ответ
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusOk, response["status"])
		assert.Equal(t, "100", response["balance"]) // Формат числа без лишних нулей
	})

	t.Run("Invalid address length", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		// Создаем запрос с некорректным адресом
		req := httptest.NewRequest(http.MethodGet, "/api/wallet/short_address/balance", nil)
		req = mux.SetURLVars(req, map[string]string{"address": "short_address"})
		rr := httptest.NewRecorder()

		server.GetBalanceHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, InvalidAddr, response["message"])
	})

	t.Run("Non-existent address", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		address := generateTestAddress("a")
		store.On("GetBalance", address).Return(0.0, storage.ErrAddressNotExist)

		req := httptest.NewRequest(http.MethodGet, "/api/wallet/"+address+"/balance", nil)
		req = mux.SetURLVars(req, map[string]string{"address": address})
		rr := httptest.NewRecorder()

		server.GetBalanceHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, "the address is not exists", response["message"])
	})

	t.Run("Storage error", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		address := generateTestAddress("a")
		store.On("GetBalance", address).Return(0.0, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/api/wallet/"+address+"/balance", nil)
		req = mux.SetURLVars(req, map[string]string{"address": address})
		rr := httptest.NewRecorder()

		server.GetBalanceHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, "Internal server error", response["message"])
	})
}

// Тесты для SendMoneyHandler
func TestSendMoneyHandler(t *testing.T) {
	t.Run("Successful transaction", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		fromAddr := generateTestAddress("a")
		toAddr := generateTestAddress("b")
		amount := 50.0

		store.On("SendMoney", fromAddr, toAddr, amount).Return(nil)

		reqBody := transaction.Request{
			From:   fromAddr,
			To:     toAddr,
			Amount: amount,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/send", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		server.SendMoneyHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusOk, response["status"])
	})

	t.Run("Invalid JSON body", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		req := httptest.NewRequest(http.MethodPost, "/api/send", bytes.NewReader([]byte("invalid json")))
		rr := httptest.NewRecorder()

		server.SendMoneyHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, "Invalid request body", response["message"])
	})

	t.Run("Invalid address length", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		reqBody := transaction.Request{
			From:   "short_from",
			To:     generateTestAddress("b"),
			Amount: 50.0,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/send", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		server.SendMoneyHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, InvalidAddr, response["message"])
	})

	t.Run("Non-positive amount", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		reqBody := transaction.Request{
			From:   generateTestAddress("a"),
			To:     generateTestAddress("b"),
			Amount: 0,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/send", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		server.SendMoneyHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, InvalidAmount, response["message"])
	})

	t.Run("Non-existent address", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		fromAddr := generateTestAddress("a")
		toAddr := generateTestAddress("b")
		amount := 50.0

		store.On("SendMoney", fromAddr, toAddr, amount).Return(storage.ErrAddressNotExist)

		reqBody := transaction.Request{
			From:   fromAddr,
			To:     toAddr,
			Amount: amount,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/send", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		server.SendMoneyHandler(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, "Address does not exist", response["message"])
	})

	t.Run("Insufficient funds", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		fromAddr := generateTestAddress("a")
		toAddr := generateTestAddress("b")
		amount := 50.0

		store.On("SendMoney", fromAddr, toAddr, amount).Return(storage.ErrInsufficient)

		reqBody := transaction.Request{
			From:   fromAddr,
			To:     toAddr,
			Amount: amount,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/send", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		server.SendMoneyHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, "Insufficient funds in the account", response["message"])
	})

	t.Run("Storage error", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		fromAddr := generateTestAddress("a")
		toAddr := generateTestAddress("b")
		amount := 50.0

		store.On("SendMoney", fromAddr, toAddr, amount).Return(assert.AnError)

		reqBody := transaction.Request{
			From:   fromAddr,
			To:     toAddr,
			Amount: amount,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/send", bytes.NewReader(body))
		rr := httptest.NewRecorder()

		server.SendMoneyHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, "Internal server error", response["message"])
	})
}

// Тесты для GetLastHandler
func TestGetLastHandler(t *testing.T) {
	t.Run("Successful retrieval", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		count := 1
		transactions := []transaction.Request{
			{
				Id:     1,
				From:   generateTestAddress("a"),
				To:     generateTestAddress("b"),
				Amount: 50.0,
			},
		}
		store.On("GetLast", count).Return(transactions, nil)

		req := httptest.NewRequest(http.MethodGet, "/api/transactions?count=1", nil)
		rr := httptest.NewRecorder()

		server.GetLastHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response []transaction.Request
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Len(t, response, 1)
		assert.Equal(t, transactions[0], response[0])
	})

	t.Run("Invalid count", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		req := httptest.NewRequest(http.MethodGet, "/api/transactions?count=invalid", nil)
		rr := httptest.NewRecorder()

		server.GetLastHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, InvalidCount, response["message"])
	})

	t.Run("Non-positive count", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		req := httptest.NewRequest(http.MethodGet, "/api/transactions?count=0", nil)
		rr := httptest.NewRecorder()

		server.GetLastHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, InvalidCount, response["message"])
	})

	t.Run("Storage error", func(t *testing.T) {
		store := &mockStorage{}
		server := setupTestServer(t, store)

		count := 1
		store.On("GetLast", count).Return([]transaction.Request{}, assert.AnError)

		req := httptest.NewRequest(http.MethodGet, "/api/transactions?count=1", nil)
		rr := httptest.NewRecorder()

		server.GetLastHandler(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response map[string]string
		err := json.NewDecoder(rr.Body).Decode(&response)
		assert.NoError(t, err)
		assert.Equal(t, StatusError, response["status"])
		assert.Equal(t, "Internal server error", response["message"])
	})
}
