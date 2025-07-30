package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Petro-vich/transaction_processing_go/internal/lib/logger/sl"
)

type Request struct {
	From   string  `json:"from"`
	To     string  `json:"to"`
	Amount float64 `json:"amount"`
}

func (sr *Server) SendMoneyHandler(w http.ResponseWriter, r *http.Request) {
	//const op = "httpserver.SendMoneyHandler"

	var req Request

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
