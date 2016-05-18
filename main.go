package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"net/http"
)

type HttpErrorResponse struct {
	status int
	msg    string
}

func WriteErrorResponse(w http.ResponseWriter, resp HttpErrorResponse) {
	http.Error(w, fmt.Sprintf(`{"status":%d,"msg":"%s"}`, resp.status, resp.msg), resp.status)
}

func HomeHandler(resp http.ResponseWriter, req *http.Request) {
}

var SessionStore *sessions.CookieStore

func setupSessions() {
	// Use a 32 byte key to select AES-256
	SessionStore = sessions.NewCookieStore(securecookie.GenerateRandomKey(32))
	SessionStore.Options = &sessions.Options{
		Path: "/",
		//Domain:   "http://mydomain.com/",
		MaxAge:   3600 * 4,
		Secure:   true,
		HttpOnly: true,
	}
}

func main() {
	setupSessions()

	r := mux.NewRouter()
	r.HandleFunc("/session/{action}", SessionHandler)
	r.HandleFunc("/", HomeHandler)
	http.Handle("/", r)

	http.ListenAndServe(":8888", nil)
}
