package main

import (
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/context"
	HttpResponse "github.com/irrenhaus/pushmearound_server/http"
	"github.com/irrenhaus/pushmearound_server/models"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	SessionName string = "pushmearound_session"
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
	token.Claims["exp"] = int64(time.Now().Add(time.Hour * 72).Unix())

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

func removeUserTokenFromSession(resp http.ResponseWriter, req *http.Request) {
	session, err := SessionStore.Get(req, SessionName)
	if err != nil {
		log.Println(err.Error())
	}

	session.Values[ContextKeyTokenString] = ""
	session.Options.MaxAge = -1
	session.Save(req, resp)
}

func parseUserToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return TokenVerifyKey, nil
	})

	if err == nil && !token.Valid {
		return nil, errors.New("Invalid token")
	}

	return token, err
}

func authenticateToken(tokenString string) (*jwt.Token, *models.AccessToken, error) {
	token, err := parseUserToken(tokenString)

	if err != nil {
		// Not a valid token
		log.Println(err.Error())
		return nil, nil, err
	}

	// Find the token in the DB to check if it's still valid
	accessToken := models.AccessToken{}
	if DB.Where("token = ?", tokenString).First(&accessToken).RecordNotFound() || accessToken.ID == 0 {
		// We have a valid token but it is not found in the DB
		return nil, nil, errors.New("Token not found (expired?)")
	}

	exp := token.Claims["exp"]
	if exp != nil && int64(token.Claims["exp"].(float64)) < time.Now().Unix() {
		// We have a valid token but it is expired
		deleteUserToken(tokenString)
		return nil, nil, errors.New("Token expired")
	}

	// There is a valid token which also is found in the DB
	return token, &accessToken, nil
}

func AuthMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	context.Set(r, ContextKeyAuthenticated, false)

	var token *jwt.Token
	var accessToken *models.AccessToken
	var err error
	user := models.User{}

	authHeader := r.Header.Get("Authorization")
	// Do we have an Bearer token in the Authorization header?
	if strings.Contains(strings.ToLower(authHeader), "bearer") {
		fields := strings.Split(authHeader, " ")

		if len(fields) > 1 {
			// Parse the token from the auth header
			token, accessToken, err = authenticateToken(fields[1])
		}
	}

	// Only check the cookie for a token if we are not already authenticated via
	// header
	if token == nil || accessToken == nil || err != nil {
		session, err := SessionStore.Get(r, SessionName)
		if err == nil {
			if tokenString, ok := session.Values[ContextKeyTokenString].(string); ok {
				token, accessToken, err = authenticateToken(tokenString)
			}
		}
	}

	// Check if the token was given via HTTP params
	if token == nil || accessToken == nil || err != nil {
		if r.ParseForm() == nil {
			tokenString := r.FormValue("token")
			if len(tokenString) > 0 {
				token, accessToken, err = authenticateToken(tokenString)
			}
		}
	}

	// Now set our context values
	context.Set(r, ContextKeyTokenID, nil)
	context.Set(r, ContextKeyTokenString, "")
	context.Set(r, ContextKeyUser, nil)

	if accessToken != nil {
		context.Set(r, ContextKeyTokenID, accessToken.ID)
		context.Set(r, ContextKeyTokenString, accessToken.Token)

		if DB.Model(&accessToken).Related(&user).RecordNotFound() {
			// Whupsy, could not find the user for the access token.
			// Reset all the variables to auth error
			err = fmt.Errorf("User not found: %d", accessToken.UserID)

			token = nil
			accessToken = nil
			log.Println(err.Error())
		} else {
			context.Set(r, ContextKeyUser, user)
		}
	}

	context.Set(r, ContextKeyError, err)
	context.Set(r, ContextKeyAuthenticated, (token != nil && err == nil))
	context.Set(r, ContextKeyToken, token)

	if token == nil || err != nil {
		removeUserTokenFromSession(w, r)
	}

	next(w, r)
}

func MustAuthenticateWrapper(f func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		authenticated := context.Get(r, ContextKeyAuthenticated)
		user := context.Get(r, ContextKeyUser)

		if authenticated == nil || !(authenticated.(bool)) || user == nil {
			WriteJSONResponse(w, HttpResponse.NewUnauthorized("unauthorized"))
			return
		}

		f(w, r)
	}
}

func LoginHandler(resp http.ResponseWriter, req *http.Request) {
	authenticated, ok := context.GetOk(req, ContextKeyAuthenticated)

	if ok && authenticated.(bool) {
		WriteJSONResponse(resp, HttpResponse.NewSuccess("You already are authenticated"))
		return
	}

	user, err := authenticateUser(req)
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

	session, err := SessionStore.Get(req, SessionName)
	if err != nil {
		WriteJSONResponse(resp, HttpResponse.NewInternalServerError(err.Error()))
		return
	}

	session.Values[ContextKeyTokenString] = token.Token
	session.Save(req, resp)

	fmt.Fprintf(resp, `{"access_token": "%s"}`, token.Token)
}

func LogoutHandler(resp http.ResponseWriter, req *http.Request) {
	token := context.Get(req, ContextKeyTokenString)
	deleteUserToken(token.(string))
	removeUserTokenFromSession(resp, req)
}
