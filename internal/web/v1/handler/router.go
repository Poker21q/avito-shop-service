package handler

import (
	"github.com/gorilla/mux"
	"merch/internal/web/v1/middleware"
	"net/http"
)

type Service interface {
	PurchaseService
	InfoService
	CoinService
	AuthService
}

type Logger interface {
	AuthLogger
	InfoLogger
	CoinLogger
	PurchaseLogger
}

type Router struct {
	service Service
	logger  Logger
}

func NewRouter(service Service, logger Logger, jwtSecret string) *mux.Router {
	r := mux.NewRouter()
	router := &Router{service: service, logger: logger}

	r.Use(middleware.Recovery(logger))

	r.Handle("/api/auth", http.HandlerFunc(router.authHandler)).Methods(http.MethodPost)

	authenticated := r.NewRoute().Subrouter()
	authenticated.Use(middleware.NewJWT(jwtSecret, logger).Authenticate)
	authenticated.Handle("/api/info", http.HandlerFunc(router.infoHandler)).Methods(http.MethodGet)
	authenticated.Handle("/api/sendCoin", http.HandlerFunc(router.sendCoinHandler)).Methods(http.MethodPost)
	authenticated.Handle("/api/buy/{item}", http.HandlerFunc(router.buyItemHandler)).Methods(http.MethodGet)

	return r
}

func (r *Router) authHandler(w http.ResponseWriter, req *http.Request) {
	h := NewAuthHandler(r.service, r.logger)
	h.Handle(w, req)
}

func (r *Router) infoHandler(w http.ResponseWriter, req *http.Request) {
	h := NewInfoHandler(r.service, r.logger)
	h.Handle(w, req)
}

func (r *Router) sendCoinHandler(w http.ResponseWriter, req *http.Request) {
	h := NewSendCoinHandler(r.service, r.logger)
	h.Handle(w, req)
}

func (r *Router) buyItemHandler(w http.ResponseWriter, req *http.Request) {
	h := NewBuyItemHandler(r.service, r.logger)
	h.Handle(w, req)
}
