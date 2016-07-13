package main

import (
	"database/sql"
	"html"
	"html/template"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dchest/uniuri"
	"github.com/gorilla/mux"
)

const (
	PORT   = ":8080"
	LENGTH = 12
)

var templates = template.Must(template.ParseFiles("static/index.html", "static/login.html", "static/register.html", "static/todo.html", "static/edit.html", "static/add.html"))

type User struct {
	ID       int
	Email    string
	Password string
}

type Tasks struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Task    string `json:"task"`
	Created string `json:"created"`
	DueDate string `json:"duedate"`
	Email   string `json:"email"`
}

type Page struct {
	Tasks []Tasks `json:"tasks"`
}

func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func genName() string {
	name := uniuri.NewLen(LENGTH)
	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)

	_, err := db.QueryRow("select name from tasks where name=?", name)
	if err != sql.ErrNoRows {
		genName()
	}
	checkErr(err)
	return name
}
func loggedIn(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	cookieValue := make(map[string]string)
	if err != nil {
		return false
	}
	err := cookieHandler.Decode("session", cookie.Value, &cookieValue)
	if err != nil {
		return false
	}
	email := cookieValue["email"]
	if email != "" {
		return false
	}
	return true

}

func getEmail(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session")
	cookieValue := make(map[string]string)
	if err != nil {
		return "", err
	}
	err := cookieHandler.Decode("session", cookie.Value, &cookieValue)
	if err != nil {
		return "", err
	}
	return cookieValue["email"], nil

}

func rootHandler(w http.ResponseWriter, r *http.Request) {

	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)

	query, err := db.Prepare("select name, title, task, created, duedate from tasks where email=? order by duedate desc")
	checkErr(err)

	email := getEmail(r)
	rows, err := query.Exec(email)
	b := Page{Tasks: Tasks{}}

	for rows.Next() {
		res := Tasks{}
		rows.scan(&res.Name, &res.Title, &res.Task, &res.Created, &res.DueDate)

		b.Tasks = append(b.Tasks, res)
	}

	checkErr(err)

	err = templates.ExecuteTemplate(w, "index.html", &b)
	checkErr(err)

}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todo := vars["id"]
	p := Page{}
	err := templates.ExecuteTemplate(w, "todo.html", &p)
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
		name := genName()

		db, err := sql.Open("mysql", DATABASE)
		checkErr(err)
		query, err := db.Prepare("insert into tasks(name, title, task, duedate, created, email)")
		err := query.Exec(name, html.EscapeString(title), html.EscapeString(task), html.EscapeString(duedate), time.Now().Format("2016-02-01 15:12:52"), email)
		checkErr(err)

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
	switch r.Method {
	case "GET":
		err := templates.ExecuteTemplate(w, "deluser.html", "")
		checkErr(err)
	case "POST", "DEL":
		pass := r.FormValue("password")

		db, err := sql.Open("mysql", DATABASE)
		checkErr(err)

		defer db.Close()

		query, err := db.Prepare("delete from users where email=? and password=?")
		checkErr(err)

		email := getEmail(r)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		checkErr(err)

		_, err = query.Exec(email, hashedPassword)
		checkErr(err)

	}

}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		err := templates.ExecuteTemplate(w, "login.html", "")
		checkErr(err)
	case "POST":
		email := r.FormValue("email")
		pass := r.FormValue("pass")
		// open db connection
		db, err := sql.Open("mysql", DATABASE)
		checkErr(err)

		defer db.Close()

		// declare variables for database results
		var hashedPassword []byte
		// read hashedPassword, name and level into variables
		err = db.QueryRow("select password from users where email=?", html.EscapeString(email)).Scan(&hashedPassword)
		if err == sql.ErrNoRows {
			http.Redirect(w, r, "/login", 303)
			return
		}
		checkErr(err)

		// compare bcrypt hash to userinput password
		err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
		if err == nil {
			// prepare cookie
			value := map[string]string{
				"email": email,
			}
			// encode variables into cookie
			if encoded, err := cookieHandler.Encode("session", value); err == nil {
				cookie := &http.Cookie{
					Name:  "session",
					Value: encoded,
					Path:  "/",
				}
				// set user cookie
				http.SetCookie(w, cookie)
			}
			// Redirect to home page
			http.Redirect(w, r, "/", 302)
		}
		// Redirect to login page
		http.Redirect(w, r, "/login", 302)

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
		db, err := sql.Open("mysql", DATABASE)
		checkErr(err)

		defer db.Close()
		_, err := db.QueryRow("select email from users where email=?", html.EscapeString(email))
		checkErr(err)
		if err == sql.ErrNoRows {
			query, err := db.Prepare("INSERT into users(email, password) values(?, ?)")
			checkErr(err)
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
			checkErr(err)

			_, err = query.Exec(html.EscapeString(email), hashedPassword)
			checkErr(err)
			http.Redirect(w, r, "/login", 302)

		}
		http.Redirect(w, r, "/register", 302)

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

	router.HandleFunc("/todo/finish/{id}", finishHandler)

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
