package transaction

import "time"

type Request struct {
	Id         int       `json:"id"`
	From       string    `json:"from"`
	To         string    `json:"to"`
	Amount     float64   `json:"amount"`
	Created_at time.Time `json:"created_at"`
}
