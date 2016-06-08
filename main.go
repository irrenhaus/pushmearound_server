package main

import (
	"crypto/rsa"
	"database/sql"
	"io/ioutil"
	"net/http"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/irrenhaus/pushmearound_server/models"
	_ "github.com/lib/pq"
	_ "github.com/mattes/migrate/driver/postgres"
	"github.com/mattes/migrate/migrate"
)

const (
	privKeyPath = "keys/app.rsa"     // openssl genrsa -out app.rsa keysize
	pubKeyPath  = "keys/app.rsa.pub" // openssl rsa -in app.rsa -pubout > app.rsa.pub
)

var (
	TokenVerifyKey *rsa.PublicKey
	TokenSignKey   *rsa.PrivateKey
)

func HomeHandler(resp http.ResponseWriter, req *http.Request) {
}

var SessionStore *sessions.CookieStore
var DB *sql.DB

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
	admin, err := models.FindUserByLogin(DB, "admin")
	if err != nil {
		if err != sql.ErrNoRows {
			log.Fatal(err)
		}

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

		err := admin.Create(DB)
		if err != nil {
			log.WithFields(log.Fields{"user": "admin", "error": err}).Fatal("Error inserting user seed")
		}

		log.WithFields(log.Fields{"user": "admin"}).Info("Inserted seed")
	}
}

func main() {
	setupSessions()

	allErrors, ok := migrate.UpSync("postgres://localhost/pushmearound?user=pushmearound&sslmode=disable&password=pushmearound", "./migrations")
	if !ok {
		for _, e := range allErrors {
			log.Error(e)
		}
		log.Fatal("Failed to migrate the database")
	}

	var err error
	DB, err = sql.Open("postgres", "host=localhost user=pushmearound dbname=pushmearound sslmode=disable password=pushmearound")
	if err != nil {
		log.Fatal(err)
	}

	seed()

	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)

	r.HandleFunc("/session/login", LoginHandler)
	r.HandleFunc("/session/logout", MustAuthenticateWrapper(LogoutHandler))

	onlyPOSTRouter := r.Methods("POST").Subrouter()
	onlyPOSTRouter.HandleFunc("/device/create", MustAuthenticateWrapper(DeviceCreateHandler))
	onlyPOSTRouter.HandleFunc("/device/options", MustAuthenticateWrapper(DeviceOptionsHandler))

	// Sending messages needs to happen as multipart/form-data
	onlyPOSTRouter.Headers("Content-Type", "multipart/form-data").Subrouter().HandleFunc("/msg/send", MustAuthenticateWrapper(SendMessageHandler))

	onlyGETRouter := r.Methods("GET").Subrouter()
	onlyGETRouter.HandleFunc("/msg/unread", MustAuthenticateWrapper(UnreadMessageHandler))

	n := negroni.Classic()
	n.Use(negroni.HandlerFunc(AuthMiddleware))
	n.UseHandler(r)

	log.WithFields(log.Fields{"host": "localhost", "port": "8888"}).Info("Now listening to http://localhost:8888")

	http.ListenAndServe(":8888", n)
}
