package httpserver

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Petro-vich/transaction_processing_go/internal/lib/logger/sl"
	"github.com/Petro-vich/transaction_processing_go/internal/models/transaction"
	"github.com/Petro-vich/transaction_processing_go/internal/storage"
	"github.com/gorilla/mux"
)

const (
	InvalidAddr   = "invalid wallet address"
	EmptyRequest  = "empty request"
	InvalidAmount = "amount must be positive"
	InvalidCount  = "count must be a positive integer"
)

const (
	StatusOk    = "OK"
	StatusError = "Error"
)

func sendError(w http.ResponseWriter, err string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  StatusError,
		"message": err,
	})
}

func (sr *Server) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	const op = "handlers.getbalancehandler"

	arg := mux.Vars(r)
	adr := arg["address"]

	if len(adr) != 64 {
		sendError(w, InvalidAddr, http.StatusBadRequest)
		sr.log.Info("Invalid wallet address length", slog.String("op", op), slog.String("address", adr))
		return
	}

	balance, err := sr.storage.GetBalance(adr)
	if err == storage.ErrAddressNotExist {
		sr.log.Info("address not found", slog.String("op", op), slog.String("address", adr))
		sendError(w, "the address is not exists", http.StatusBadRequest)
		return
	} else if err != nil {
		sendError(w, "Internal server error", http.StatusInternalServerError)
		sr.log.Error("Error get balance", slog.String("op", op), sl.Err(err))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  StatusOk,
		"balance": strconv.FormatFloat(balance, 'g', -1, 64),
	})
}

func (sr *Server) SendMoneyHandler(w http.ResponseWriter, r *http.Request) {
	const op = "httpserver.SendMoneyHandler"

	var req transaction.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sr.log.Info("error decode", slog.String("op", op), sl.Err(err))
		sendError(w, "Invalid request body", http.StatusBadRequest)
		sr.log.Error("Failed to decode request body", slog.String("op", op), sl.Err(err))
		return
	}

	if len(req.From) != 64 || len(req.To) != 64 {
		sendError(w, InvalidAddr, http.StatusBadRequest)
		sr.log.Error("Invalid wallet address length", slog.String("op", op))
		return
	}

	if req.Amount <= 0 {
		sendError(w, InvalidAmount, http.StatusBadRequest)
		sr.log.Info(InvalidAmount, slog.String("op", op), slog.Float64("amount", (req.Amount)))
		return
	}

	sr.log.Info("Request body decoded", slog.String("op", op))

	err := sr.storage.SendMoney(req.From, req.To, req.Amount)
	if err == storage.ErrAddressNotExist {
		sendError(w, "Address does not exist", http.StatusNotFound)
		sr.log.Info("Address not found")
		return
	} else if err == storage.ErrInsufficient {
		sendError(w, "Insufficient funds in the account", http.StatusBadRequest)
		sr.log.Info("Insufficient funds", slog.String("op", op))
		return
	} else if err != nil {
		sr.log.Error("Failed to send money", slog.String("op", op), sl.Err(err))
		sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sr.log.Info("Transaction completed successfully", slog.String("op", op))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": StatusOk,
	})
}

func (sr *Server) GetLastHandler(w http.ResponseWriter, r *http.Request) {
	const op = "httpserver.GetLastHandler"

	query := r.URL.Query()
	strCount := query.Get("count")

	count, err := strconv.Atoi(strCount)
	if err != nil || count <= 0 {
		sendError(w, InvalidCount, http.StatusBadRequest)
		sr.log.Error(InvalidCount, slog.String("count", strCount))
		return
	}

	transactions, err := sr.storage.GetLast(count)
	if err != nil {
		sr.log.Error("Couldn't get the latest transactions", slog.String("op", op), sl.Err(err))
		sendError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sr.log.Info("Retrieved transactions", slog.String("op", op), slog.Int("count", count))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}
