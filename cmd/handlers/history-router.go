package handlers

import (
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"net/http"
)

func HistoryRouter(log *zerolog.Logger, myReqList *MyRequestList, proxy *Proxy) http.Handler {
	history := History{Log: log, MyReqList: myReqList, Proxy: proxy}
	router := mux.NewRouter()
	router.HandleFunc("/requests", history.GetLastRequestsList).Methods("GET").Queries("limit", "{limit}", "offset", "{offset}")
	router.HandleFunc("/response/{id}", history.GetResponse).Methods("GET")
	return router
}
