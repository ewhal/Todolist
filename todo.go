package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
)

const (
	PORT   = ":8080"
	LENGTH = 12
)

var templates = template.Must(template.ParseFiles("static/index.html", "static/login.html", "static/register.html", "static/todo.html", "static/edit.html", "static/add.html"))

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func genName() string {
	name := uniuri.NewLen(LENGTH)

	return name
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "index.html", "")
	checkErr(err)

}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todo := vars["id"]
	err := templates.ExecuteTemplate(w, "todo.html", "")
	checkErr(err)

}

func addHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		err := templates.ExecuteTemplate(w, "add.html", "")
		checkErr(err)

	case "POST":
		title := r.FormValue("title")
		task := r.FormValue("task")
		duedate := r.FormValue("duedate")

	}

}

func editHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todo := vars["id"]

}

func delHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todo := vars["id"]

}

func finishHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todo := vars["id"]

}

func userHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

}

func userDelHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		err := templates.ExecuteTemplate(w, "login.html", "")
		checkErr(err)
	case "POST":
		email := r.FormValue("email")
		pass := r.FormValue("pass")
	}

}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		err := templates.ExecuteTemplate(w, "login.html", "")
		checkErr(err)
	case "POST":
		email := r.FormValue("email")
		pass := r.FormValue("pass")
	}

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", 301)

}

func resetHandler(w http.ResponseWriter, r *http.Request) {

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
