package main

import (
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"html/template"
	"net/http"
)

// cookie handling

var cookieHandler = securecookie.New(
	securecookie.GenerateRandomKey(64),
	securecookie.GenerateRandomKey(32))

func getUserName(request *http.Request) (userName string) {
	if cookie, err := request.Cookie("session"); err == nil {
		cookieValue := make(map[string]string)
		if err = cookieHandler.Decode("session", cookie.Value, &cookieValue); err == nil {
			userName = cookieValue["name"]
		}
	}
	return userName
}

func setSession(userName string, response http.ResponseWriter) {
	value := map[string]string{
		"name": userName,
	}
	if encoded, err := cookieHandler.Encode("session", value); err == nil {
		cookie := &http.Cookie{
			Name:  "session",
			Value: encoded,
			Path:  "/",
		}
		http.SetCookie(response, cookie)
	}
}

func clearSession(response http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	}
	http.SetCookie(response, cookie)
}

// login handler

func loginHandler(response http.ResponseWriter, request *http.Request) {
	name := request.FormValue("name")
	pass := request.FormValue("password")
	redirectTarget := "/"
	if name == "admin" && pass != "" {
		// .. check credentials ..
		setSession(name, response)
		redirectTarget = "/internal"
	} else if name == "worker" && pass != "" {
		setSession(name, response)
		redirectTarget = "/internal2"
	}
	http.Redirect(response, request, redirectTarget, 302)
}

// logout handler

func logoutHandler(response http.ResponseWriter, request *http.Request) {
	clearSession(response)
	http.Redirect(response, request, "/", 302)
}

/*
const internalPage = `
<h1>Internal</h1>
<hr>
<small>User: %s</small>
<form method="post" action="/logout">
    <button type="submit">Logout</button>
</form>
`
const internalPage2 = `
<h1>Worker</h1>
<hr>
<small>User: %s</small>
<form method="post" action="/logout">
    <button type="submit">Logout</button>
</form>
`

func internalPageHandler(response http.ResponseWriter, request *http.Request) {
	userName := getUserName(request)
	if userName != "" {
		fmt.Fprintf(response, internalPage, userName)
	} else {
		http.Redirect(response, request, "/", 302)
	}
}
*/

var tmpl *template.Template

func init() {
	tmpl = template.Must(template.ParseGlob("templates/*.html"))
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "login_new.html", nil)
}

// internal page
func internalPageWorker(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "worker.html", nil)
}

func internalPageAdmin(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "admin.html", nil)
}

// server main method
var router = mux.NewRouter()

func main() {

	fs := http.FileServer((http.Dir("assets")))
	http.Handle("/assets/", http.StripPrefix("/assets", fs))
	router.HandleFunc("/", homeHandler)

	//router.HandleFunc("/", indexPageHandler)
	router.HandleFunc("/internal2", internalPageWorker)
	router.HandleFunc("/internal", internalPageAdmin)

	router.HandleFunc("/login", loginHandler).Methods("POST")
	router.HandleFunc("/logout", logoutHandler).Methods("POST")

	http.Handle("/", router)
	http.ListenAndServe(":80", nil)

}
