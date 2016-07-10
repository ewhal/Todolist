package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

const (
	PORT = ":8080"
)

func rootHandler(w http.responsewriter, r *http.request) {

}

func todoHandler(w http.responsewriter, r *http.request) {

}

func addHandler(w http.responsewriter, r *http.request) {

}

func editHandler(w http.responsewriter, r *http.request) {

}

func delHandler(w http.responsewriter, r *http.request) {

}

func finishHandler(w http.responsewriter, r *http.request) {

}

func userHandler(w http.responsewriter, r *http.request) {

}

func userDelHandler(w http.responsewriter, r *http.request) {

}

func loginHandler(w http.responsewriter, r *http.request) {

}

func registerHandler(w http.responsewriter, r *http.request) {

}

func logoutHandler(w http.responsewriter, r *http.request) {

}

func resetHandler(w http.responsewriter, r *http.request) {

}

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/", rootHandler)

	router.HandleFunc("/todo", todoHandler)
	router.HandleFunc("/todo/{id}", todoHandler)
	router.HandleFunc("/todo/add", addHandler)
	router.HandleFunc("/todo/edit/{id}", editHandler)
	router.HandleFunc("/todo/del/{id}", delHandler)

	router.HandleFunc("/finish/{id}", finishHandler)

	router.HandleFunc("/user", userHandler)
	router.HandleFunc("/user/{id}", userHandler)
	router.HandleFunc("/user/del/{id}", userDelHandler)

	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/logout", logoutHandler)
	router.HandleFunc("/resetpass", resetHandler)
	err := http.ListenAndServe(PORT, router)
	if err != nil {
		log.Fatal(err)
	}

}
