package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func SessionHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	action, exists := vars["action"]

	if !exists {
		WriteErrorResponse(resp, HttpErrorResponse{http.StatusBadRequest, "No session action specified"})
		return
	}

	session, err := SessionStore.Get(req, "pushmearound_session")
	if err != nil {
		WriteErrorResponse(resp, HttpErrorResponse{http.StatusInternalServerError, err.Error()})
		return
	}

	switch action {
	case "login":
		err = req.ParseForm()
		if err != nil {
			WriteErrorResponse(resp, HttpErrorResponse{http.StatusBadRequest, "Could not parse form data"})
			return
		}

		username := req.FormValue("username")
		password := req.FormValue("password")

		if len(username) < 3 || len(password) < 6 {
			WriteErrorResponse(resp, HttpErrorResponse{http.StatusBadRequest, "username/password too short"})
			return
		}

		session.Save(req, resp)
	case "logout":
		session.Options.MaxAge = -1
		session.Save(req, resp)
	default:
		WriteErrorResponse(resp, HttpErrorResponse{http.StatusBadRequest, fmt.Sprintf("Unknown session action '%s'", action)})
		return
	}
}
