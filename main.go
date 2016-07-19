// package main is a robust modern todolist service
package main

import (
	"database/sql"
	"encoding/json"
	"html"
	"html/template"
	"log"
	"net/http"
	"os"

	// uniuri for random string generation
	"github.com/dchest/uniuri"
	// mysql driver
	_ "github.com/go-sql-driver/mysql"
	// mux routing
	"github.com/gorilla/mux"
	// securecookie for cookie handling
	"github.com/gorilla/securecookie"
	// bcrypt for password hashing
	"golang.org/x/crypto/bcrypt"
)

type Configuration struct {
	// PORT for golang to listen on
	Port string
	// LENGTH todo name length
	Length int
	// USERNAME database username
	Username string
	// PASS database password
	Password string
	// NAME database name
	Name string
}

var configuration Configuration

// DATABASE connection String
var DATABASE string

var templates = template.Must(template.ParseFiles("static/index.html", "static/login.html", "static/register.html", "static/todo.html", "static/task.html"))

// generate new random cookie keys
var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32),
)

// User struct
type User struct {
	ID       int
	Email    string
	Password string
}

// Tasks todo tasks struct
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

// Page []Tasks struct
type Page struct {
	Tasks []Tasks `json:"tasks"`
}

// Cal fullcalendar struct
type Cal struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Start       string `json:"start"`
	End         string `json:"end"`
	Allday      bool   `json:"allDay"`
	URL         string `json:"url"`
}

// checkErr logger
func checkErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

// genName Random name geneation function
func genName() string {
	// use uniuri to generate random string
	name := uniuri.NewLen(configuration.Length)

	// open db connection
	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)

	_, err = db.Query("select name from tasks where name=?", name)
	db.Close()
	// if name exists in db call genName again
	if err != sql.ErrNoRows {
		genName()
	}
	checkErr(err)
	// return random string
	return name
}

// loggedIn returns true if cookie exists
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

// getEmail returns the users email address from cookie
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

// rootHandler root page handler
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// check if user is logged in, if not redirect to login page
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}

	err := templates.ExecuteTemplate(w, "index.html", "")
	checkErr(err)
}

// calHandler generates json string that is fullcalendar compatible
func calHandler(w http.ResponseWriter, r *http.Request) {
	// check if user is logged in, if not redirect to login page
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

// taskHandler
func taskHandler(w http.ResponseWriter, r *http.Request) {
	// check if user is logged in, if not redirect to login page
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
	// get todo name
	vars := mux.Vars(r)
	todo := vars["id"]

	// open db connection
	db, err := sql.Open("mysql", DATABASE)
	checkErr(err)
	defer db.Close()

	// query if todo is public
	rows, err := db.Query("select public from tasks where name=?", html.EscapeString(todo))
	checkErr(err)
	var public bool
	for rows.Next() {
		rows.Scan(&public)
	}

	// check if user is logged in or if todo is public, if not redirect to login page
	if loggedIn(r) != true || public == false {
		http.Redirect(w, r, "/login", 302)
	}
	p := Tasks{}
	// query todo information from db
	query, err := db.Query("select title, task, duedate, created, completed, allday from tasks where name=?", html.EscapeString(todo))
	checkErr(err)

	for query.Next() {
		query.Scan(&p.Title, &p.Task, &p.DueDate, &p.Created, &p.Completed, &p.Allday)
	}

	// Execute task template
	err = templates.ExecuteTemplate(w, "task.html", &p)
	checkErr(err)

}

// apitodoHandler
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

	// check if user is logged in or if todo is public, if not redirect to login page
	if loggedIn(r) != true || public == false {
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

// addHandler
func addHandler(w http.ResponseWriter, r *http.Request) {
	// check if user is logged if not redirect to login page
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

// editHandler
func editHandler(w http.ResponseWriter, r *http.Request) {
	// check if user is logged if not redirect to login page
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
	_, err = query.Exec(html.EscapeString(title), html.EscapeString(task), html.EscapeString(duedate), html.EscapeString(public), html.EscapeString(created), html.EscapeString(allday), html.EscapeString(todo), email)
	checkErr(err)

}

// delHandler
func delHandler(w http.ResponseWriter, r *http.Request) {
	// check if user is logged if not redirect to login page
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

// finishHandler
func finishHandler(w http.ResponseWriter, r *http.Request) {
	// check if user is logged if not redirect to login page
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

// userHandler
func userHandler(w http.ResponseWriter, r *http.Request) {
	// check if user is logged if not redirect to login page
	if loggedIn(r) != true {
		http.Redirect(w, r, "/login", 302)
	}

}

// userDelHandler
func userDelHandler(w http.ResponseWriter, r *http.Request) {
	// check if user is logged if not redirect to login page
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

// loginHandler
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

// registerHandler
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

// logoutHandler destroys cookie data and redirects to root
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

// resetHandler is meant to handle resetting the users password if forgotten
func resetHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	file, err := os.Open("config.json")
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err = decoder.Decode(&configuration)
	if err != nil {
		panic(err)
	}

	DATABASE = configuration.Username + ":" + configuration.Password + "@/" + configuration.Name + "?charset=utf8"
	// create new mux router
	router := mux.NewRouter()

	// basic handlers
	router.HandleFunc("/", rootHandler)
	router.HandleFunc("/todo", taskHandler)
	router.HandleFunc("/todo/{id}", todoHandler).Methods("GET")

	// api handlers
	router.HandleFunc("/api/cal", addHandler).Methods("POST")
	router.HandleFunc("/api/cal", calHandler).Methods("GET")
	router.HandleFunc("/api/cal/{id}", apitodoHandler).Methods("GET")
	router.HandleFunc("/api/cal/{id}", editHandler).Methods("PUT", "POST")
	router.HandleFunc("/api/cal/{id}", delHandler).Methods("DELETE")

	router.HandleFunc("/finish/{id}", finishHandler).Methods("POST")

	// user handlers
	router.HandleFunc("/user", userHandler)
	router.HandleFunc("/user/del", userDelHandler)

	// account handlers
	router.HandleFunc("/register", registerHandler)
	router.HandleFunc("/login", loginHandler)
	router.HandleFunc("/logout", logoutHandler)
	router.HandleFunc("/resetpass", resetHandler)
	// ListenAndServe on PORT with router
	err = http.ListenAndServe(configuration.Port, router)
	if err != nil {
		log.Fatal(err)
	}

}
