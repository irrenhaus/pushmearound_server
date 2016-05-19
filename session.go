package main

import (
	"errors"
	"fmt"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	HttpResponse "github.com/irrenhaus/pushmearound_server/http"
	"github.com/irrenhaus/pushmearound_server/models"
	"log"
	"net/http"
	"time"
)

func authenticateUser(req *http.Request) (*models.User, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, errors.New("Could not parse form data")
	}

	username := req.FormValue("username")
	password := req.FormValue("password")

	if len(username) < 3 || len(password) < 6 {
		return nil, errors.New("username/password too short")
	}

	user := models.User{}
	if DB.Where("username = ?", username).Or("email = ?", username).First(&user).RecordNotFound() {
		return nil, errors.New(fmt.Sprintf("No such user: '%s'", username))
	}

	if user.ComparePassword(password) != nil {
		return nil, errors.New("Please check your username and password.")
	}

	return &user, nil
}

func createUserToken(user *models.User) (*models.AccessToken, error) {
	token := jwt.New(jwt.SigningMethodRS256)

	token.Claims["user"] = user.Username
	token.Claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	tokenString, err := token.SignedString(TokenSignKey)
	if err != nil {
		log.Println("ERROR Could not created signed token", err.Error())
		return nil, errors.New("Could not create signed token")
	}

	log.Println(tokenString)

	accessToken := models.AccessToken{
		UserID: user.ID,
		Token:  tokenString,
	}

	if err := DB.Model(user).Association("Tokens").Append(accessToken).Error; err != nil {
		log.Println("ERROR Could not append access token", err.Error())
		return nil, errors.New(fmt.Sprintf("Could not append new access token for user %s", user.Username))
	}

	return &accessToken, nil
}

func deleteUserToken(token string) {
	accessToken := models.AccessToken{}
	if DB.Where("token = ?", token).First(&accessToken).RecordNotFound() || accessToken.ID == 0 {
		return
	}

	DB.Delete(&accessToken)
}

func SessionHandler(resp http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	action, exists := vars["action"]

	if !exists {
		WriteJSONResponse(resp, HttpResponse.NewBadRequest("No session action specified"))
		return
	}

	session, err := SessionStore.Get(req, "pushmearound_session")
	if err != nil {
		WriteJSONResponse(resp, HttpResponse.NewInternalServerError(err.Error()))
		return
	}

	switch action {
	case "login":
		var user *models.User
		user, err = authenticateUser(req)
		if err != nil {
			WriteJSONResponse(resp, HttpResponse.NewBadRequest(err.Error()))
			return
		}

		var token *models.AccessToken
		token, err = createUserToken(user)
		if err != nil {
			WriteJSONResponse(resp, HttpResponse.NewBadRequest(err.Error()))
			return
		}

		session.Values["token"] = token.Token
		session.Save(req, resp)

		fmt.Fprintf(resp, `{"access_token": "%s"}`, token.Token)
	case "logout":
		deleteUserToken(session.Values["token"].(string))

		session.Values["token"] = ""
		session.Options.MaxAge = -1
		session.Save(req, resp)
	default:
		WriteJSONResponse(resp, HttpResponse.NewBadRequest(fmt.Sprintf("Unknown session action '%s'", action)))
		return
	}
}
