package main

import (
    "github.com/gorilla/mux"
    "github.com/warneki/spaced/server/database"
    "net/http"
)

const (
    STATIC_DIR = "./server/public/"
)

func Router() *mux.Router {

    router := mux.NewRouter()

    router.HandleFunc("/api/today", database.GetDataForToday).Methods("GET")

    router.HandleFunc("/sessions", database.ReturnOptions).Methods("OPTIONS")
    router.HandleFunc("/sessions", database.AddNewSession).Methods("POST")

    router.PathPrefix("/").Handler(http.FileServer(http.Dir(STATIC_DIR)))

    return router
}
