package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Petro-vich/transaction_processing_go/internal/lib/logger/sl"
	"github.com/Petro-vich/transaction_processing_go/internal/models/transaction"
	"github.com/gorilla/mux"
)

func (sr *Server) GetBalanceHandler(w http.ResponseWriter, r *http.Request) {
	//const op = "handlers.getbalancehandler"
	arg := mux.Vars(r)
	adr := arg["address"]

	balance, err := sr.storage.GetBalance(adr)
	if err != nil {
		sr.log.Error("Ошибка получения баланса", sl.Err(err))
	}

	fmt.Println(balance)
}

func (sr *Server) SendMoneyHandler(w http.ResponseWriter, r *http.Request) {
	//const op = "httpserver.SendMoneyHandler"

	var req transaction.Request

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Fprintf(w, "Ошибка декода") //TODO:
		sr.log.Error("error decode", sl.Err(err))
		return
	}
	sr.log.Info("request body decoded")

	err := sr.storage.SendMoney(req.From, req.To, req.Amount)
	if err != nil {
		fmt.Println(err)
	}
	sr.log.Info("Отправка завершена успешно")
}

func (sr *Server) GetLastHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	strCount := query.Get("count")
	count, err := strconv.Atoi(strCount)
	if err != nil {
		sr.log.Debug("atrcov.atoi error", sl.Err(err))
	}

	LastOp, err := sr.storage.GetLast(count)
	if err != nil {
		sr.log.Debug("sr.storage.Getlast error", sl.Err(err))
	}

	_ = LastOp

}
