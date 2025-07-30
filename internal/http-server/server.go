package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/Petro-vich/transaction_processing_go/internal/config"
	"github.com/Petro-vich/transaction_processing_go/internal/storage/sqlite"
	"github.com/gorilla/mux"
)

type Server struct {
	storage *sqlite.Storage
	config  *config.Config
	router  *mux.Router
	log     *slog.Logger
}

func New(storage *sqlite.Storage, config *config.Config, log *slog.Logger) *Server {
	serv := Server{
		storage: storage,
		config:  config,
		router:  mux.NewRouter(),
		log:     log,
	}
	serv.routes()
	return &serv
}

func (sr *Server) Start() error {
	return http.ListenAndServe(sr.config.Address, sr.router)
}

func (sr *Server) routes() {
	sr.router.HandleFunc("/api/send", sr.SendMoneyHandler)
	
}
