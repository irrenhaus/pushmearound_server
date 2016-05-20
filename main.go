package main

import (
	"crypto/rsa"
	"encoding/json"
	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	HttpResponse "github.com/irrenhaus/pushmearound_server/http"
	"github.com/irrenhaus/pushmearound_server/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	privKeyPath = "keys/app.rsa"     // openssl genrsa -out app.rsa keysize
	pubKeyPath  = "keys/app.rsa.pub" // openssl rsa -in app.rsa -pubout > app.rsa.pub
)

var (
	TokenVerifyKey *rsa.PublicKey
	TokenSignKey   *rsa.PrivateKey
)

func WriteJSONResponse(w http.ResponseWriter, resp HttpResponse.HttpResponse) {
	jsonContent, _ := json.Marshal(&resp)
	http.Error(w, string(jsonContent), resp.Status)
}

func HomeHandler(resp http.ResponseWriter, req *http.Request) {
}

var SessionStore *sessions.CookieStore
var DB *gorm.DB

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

	signBytes, err := ioutil.ReadFile(privKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	TokenSignKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}

	verifyBytes, err := ioutil.ReadFile(pubKeyPath)
	if err != nil {
		log.Fatal(err)
	}

	TokenVerifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatal(err)
	}
}

func seed() {
	admin := models.User{}
	if DB.Where("username = ?", "admin").First(&admin).RecordNotFound() {
		admin = models.User{
			Username:       "admin",
			FirstName:      "Nils",
			LastName:       "Hesse",
			Email:          "nphesse@gmail.com",
			EmailConfirmed: true,
			Password:       "",
			LastSignInAt:   time.Now(),
		}

		admin.SetPassword("lalala")

		DB.Create(&admin)
	}
}

func main() {
	setupSessions()

	var err error
	DB, err = gorm.Open("postgres", "host=localhost user=pushmearound dbname=pushmearound sslmode=disable password=pushmearound")
	if err != nil {
		log.Fatal(err)
	}

	DB.AutoMigrate(&models.User{}, &models.AccessToken{}, &models.Device{})

	if DB.Error != nil {
		log.Fatal("AutoMigrate failed", DB.Error)
	}

	seed()

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/session/logout", MustAuthenticateWrapper(LogoutHandler))
	r.HandleFunc("/session/login", LoginHandler)

	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(AuthMiddleware))
	n.UseHandler(r)

	http.ListenAndServe(":8888", n)
}
