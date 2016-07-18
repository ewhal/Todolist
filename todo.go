package main

import (
	"database/sql"
	"encoding/json"
	"html"
	"html/template"
	"log"
	"net/http"

	"github.com/dchest/uniuri"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
)

const (
	PORT     = ":8080"
	LENGTH   = 12
	USERNAME = "root"
	// PASS database password
	PASS = ""
	// NAME database name
	NAME = ""
	// DATABASE connection String
	DATABASE = USERNAME + ":" + PASS + "@/" + NAME + "?charset=utf8"
)

var templates = template.Must(template.ParseFiles("static/index.html", "static/login.html", "static/register.html", "static/todo.html", "static/task.html"))

// generate new random cookie keys
var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

type User struct {
	ID       int
	Email    string
	Password string
}

type Tasks struct {
	Name      string `json:"name"`
	Title     string `json:"title"`
	Task      string `json:"task"`
	Created   string `json:"created"`
	DueDate   string `json:"duedate"`
	Email     string `json:"email"`
	Completed bool   `json:"completed"`
	Public    bool   `json:"public"`
	Allday    bool   `json:"allday"`
}

type Page struct {
	Tasks []Tasks `json:"tasks"`
}
type Cal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Allday      bool   `json:"allDay"`
	URL         string `json:"url"`
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

	_, err = db.Query("select name from tasks where name=?", name)
	db.Close()
	if err == sql.ErrNoRows {
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
	err = cookieHandler.Decode("session", cookie.Value, &cookieValue)
	if err != nil {
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
	err = cookieHandler.Decode("session", cookie.Value, &cookieValue)
	if err != nil {
		return "", err
	}
	return cookieValue["email"], nil

}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}

	err := templates.ExecuteTemplate(w, "index.html", "")
	checkErr(err)
}

func calHandler(w http.ResponseWriter, r *http.Request) {
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}
	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)

	email, err := getEmail(r)
	checkErr(err)

	rows, err := db.Query("select title, task, created, duedate, allday, name from tasks where email=? order by duedate asc", email)
	checkErr(err)

	b := []Cal{}

	for rows.Next() {
		res := Cal{}
		rows.Scan(&res.Title, &res.Description, &res.Start, &res.End, &res.Allday, &res.URL)

		b = append(b, res)
	}
	db.Close()

	checkErr(err)
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(b)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func taskHandler(w http.ResponseWriter, r *http.Request) {
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}

	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)

	email, err := getEmail(r)
	checkErr(err)

	rows, err := db.Query("select name, title, task, created, duedate from tasks where email=? order by duedate asc", email)
	checkErr(err)

	b := Page{Tasks: []Tasks{}}

	for rows.Next() {
		res := Tasks{}
		rows.Scan(&res.Name, &res.Title, &res.Task, &res.Created, &res.DueDate)

		b.Tasks = append(b.Tasks, res)
	}
	db.Close()

	checkErr(err)

	err = templates.ExecuteTemplate(w, "todo.html", &b)
	checkErr(err)

}
func todoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todo := vars["id"]

	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)
	defer db.Close()

	rows, err := db.Query("select public from tasks where name=?", html.EscapeString(todo))
	checkErr(err)
	var public bool
	for rows.Next() {
		rows.Scan(&public)
	}

	if loggedIn(r) != true && public == false {
		http.Redirect(w, r, "/login", 302)
	}
	p := Tasks{}
	query, err := db.Query("select title, task, duedate, created, completed, allday from tasks where name=?", html.EscapeString(todo))
	checkErr(err)

	for query.Next() {
		query.Scan(&p.Title, &p.Task, &p.DueDate, &p.Created, &p.Completed, &p.Allday)
	}

	err = templates.ExecuteTemplate(w, "task.html", &p)
	checkErr(err)

}

func apitodoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todo := vars["id"]

	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)
	defer db.Close()

	rows, err := db.Query("select public from tasks where name=?", html.EscapeString(todo))
	checkErr(err)
	var public bool
	for rows.Next() {
		rows.Scan(&public)
	}

	if loggedIn(r) != true && public == false {
		http.Redirect(w, r, "/login", 302)
	}
	p := Tasks{}
	query, err := db.Query("select title, task, duedate, created, completed, allday from tasks where name=?", html.EscapeString(todo))
	checkErr(err)

	for query.Next() {
		query.Scan(&p.Title, &p.Task, &p.DueDate, &p.Created, &p.Completed, &p.Allday)
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func addHandler(w http.ResponseWriter, r *http.Request) {
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}

	title := r.FormValue("title")
	task := r.FormValue("task")
	duedate := r.FormValue("duedate")
	public := r.FormValue("public")
	created := r.FormValue("created")
	allday := r.FormValue("allday")
	name := genName()
	email, err := getEmail(r)
	checkErr(err)

	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)
	defer db.Close()

	query, err := db.Prepare("insert into tasks(name, title, task, duedate, created, email, completed, public, allday) values(?, ?, ?, ?, ?, ?, ?, ?, ?)")
	_, err = query.Exec(name, html.EscapeString(title), html.EscapeString(task), html.EscapeString(duedate), html.EscapeString(created), email, false, html.EscapeString(public), html.EscapeString(allday))
	checkErr(err)

}

func editHandler(w http.ResponseWriter, r *http.Request) {
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}
	vars := mux.Vars(r)
	todo := vars["id"]
	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)
	defer db.Close()

	title := r.FormValue("title")
	task := r.FormValue("task")
	created := r.FormValue("created")
	duedate := r.FormValue("duedate")
	public := r.FormValue("public")
	allday := r.FormValue("allday")

	query, err := db.Prepare("update tasks set title=?, task=?, duedate=?, public=?, created=?, allday=? where name=? and email=?")
	checkErr(err)
	email, err := getEmail(r)
	checkErr(err)
	_, err = query.Exec(html.EscapeString(todo), html.EscapeString(title), html.EscapeString(task), html.EscapeString(duedate), html.EscapeString(public), html.EscapeString(created), html.EscapeString(allday), email)
	checkErr(err)

}

func delHandler(w http.ResponseWriter, r *http.Request) {
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}
	vars := mux.Vars(r)
	todo := vars["id"]

	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)
	defer db.Close()

	email, err := getEmail(r)
	checkErr(err)

	_, err = db.Query("delete from tasks where email=? and name=?", email, html.EscapeString(todo))
	checkErr(err)

}

func finishHandler(w http.ResponseWriter, r *http.Request) {
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}
	vars := mux.Vars(r)
	todo := vars["id"]
	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)
	defer db.Close()

	email, err := getEmail(r)
	checkErr(err)
	_, err = db.Query("update tasks set completed=true where email=? and name=?", email, html.EscapeString(todo))

}

func userHandler(w http.ResponseWriter, r *http.Request) {
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}

}

func userDelHandler(w http.ResponseWriter, r *http.Request) {
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}
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

		email, err := getEmail(r)
		checkErr(err)
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
		password := r.FormValue("password")
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
		err := templates.ExecuteTemplate(w, "register.html", "")
		checkErr(err)
	case "POST":
		email := r.FormValue("email")
		pass := r.FormValue("password")
		db, err := sql.Open("mysql", DATABASE)
		checkErr(err)

		defer db.Close()
		//_, err = db.Query("select email from users where email=?", html.EscapeString(email))
		//checkErr(err)
		//if err == sql.ErrNoRows {
		query, err := db.Prepare("INSERT into users(email, password) values(?, ?)")
		checkErr(err)
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
		checkErr(err)

		_, err = query.Exec(html.EscapeString(email), hashedPassword)
		checkErr(err)
		http.Redirect(w, r, "/login", 302)

		//}
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

	router.HandleFunc("/todo", taskHandler)
	router.HandleFunc("/todo/{id}", todoHandler).Methods("GET")

	router.HandleFunc("/api/cal", addHandler).Methods("POST")
	router.HandleFunc("/api/cal", calHandler).Methods("GET")
	router.HandleFunc("/api/cal/{id}", apitodoHandler).Methods("GET")
	router.HandleFunc("/api/cal/{id}", editHandler).Methods("PUT")
	router.HandleFunc("/api/cal/{id}", delHandler).Methods("DELETE")

	router.HandleFunc("/finish/{id}", finishHandler).Methods("POST")

	router.HandleFunc("/user", userHandler)
	router.HandleFunc("/user/del", userDelHandler)

	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/logout", logoutHandler)
	router.HandleFunc("/resetpass", resetHandler)
	// ListenAndServe on PORT with router
	err := http.ListenAndServe(PORT, router)
	if err != nil {
		log.Fatal(err)
	}

}
