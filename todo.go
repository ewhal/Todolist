package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	PORT = ":8080"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", roothandler)
	err := http.ListenAndServe(PORT, router)
	if err != nil {
		log.Fatal(err)
	}

}
