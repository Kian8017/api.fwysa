package main

import (
	"context"
	_ "github.com/go-kivik/couchdb"
	"github.com/go-kivik/kivik"
	"log"
	"net/http"
	"path"
)

type Server struct {
	m              *http.ServeMux
	db             *kivik.DB
	listenAddr     string
	couchdbUrl     string
	couchdbName    string
	dbAccessString string
}

func NewServer(la, cdb, dbn string) *Server {
	a := Server{}
	a.listenAddr = la
	a.couchdbUrl = cdb
	a.couchdbName = dbn
	a.dbAccessString = path.Join(a.couchdbUrl, a.couchdbName)

	sm := http.NewServeMux()
	a.m = sm

	// Required handlers
	sm.HandleFunc("/login", a.LoginHandler)

	// Helper handlers
	sm.HandleFunc("/hash", a.HashHandler)
	sm.HandleFunc("/timestamp", a.TimestampHandler)
	sm.HandleFunc("/id", a.IdHandler)
	sm.HandleFunc("/diceware", a.DicewareHandler)

	// Root handler
	sm.HandleFunc("/", a.RootHandler)

	// Connect to couchdb server
	dbclient, err := kivik.New("couch", a.couchdbUrl)
	if err != nil {
		log.Fatal("Unable to connect to database ", err)
	}
	a.db = dbclient.DB(context.TODO(), a.couchdbName)

	return &a
}

func (s *Server) Run() {
	log.Println("Server listening at", s.listenAddr)
	log.Fatal(http.ListenAndServe(s.listenAddr, s.m))
}

func (s *Server) Login(user, pass string) (AuthDocument, string) {
	// selector
	q := map[string]interface{}{
		"selector": map[string]interface{}{
			"type":     "auth",
			"username": user,
		},
	}

	rows, err := s.db.Find(context.TODO(), q)
	if err != nil {
		log.Println("Error retrieving login details", err)
		return AuthDocument{}, "internal error"
	}

	if !rows.Next() { // Either no documents or an error
		err = rows.Err()
		if err != nil {
			log.Println("Error trying to access results ", err)
			return AuthDocument{}, "internal error"
		} else {
			log.Println("No such user ", user)
			return AuthDocument{}, "no such user"
		}
	}

	var cur AuthDocument
	err = rows.ScanDoc(&cur)
	if err != nil {
		log.Println("Unable to unmarshal AuthDocument ", err)
		return AuthDocument{}, "internal error"
	}

	// Now confirm password before handing back DB details

	newHash := hash(pass)

	if newHash != cur.Password { // Incorrect password
		log.Println("Incorrect login attempt for user", cur.Username)
		return AuthDocument{}, "incorrect password"
	}

	return cur, ""
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	username, ok := query["user"]
	if !ok {
		w.Write(ErrHelper("username not provided"))
		return
	}

	pass, ok := query["pass"]
	if !ok {
		w.Write(ErrHelper("password not provided"))
		return
	}
	// We're good to go! Send them the details...

	ad, deets := s.Login(username[0], pass[0])
	if deets != "" { // Something happened
		w.Write(ErrHelper(deets))
		return
	}

	w.Write(LoginHelper(ad.Role, ad.UserID, s.dbAccessString))
}

// Helper Handlers
func (s *Server) HashHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	query := r.URL.Query()
	pass, ok := query["pass"]
	if !ok {
		w.Write(ErrHelper("pass query parameter not present"))
		return
	}

	w.Write(SimpleHelper(hash(pass[0])))
}

func (s *Server) TimestampHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(SimpleHelper(genTimestamp()))
}

func (s *Server) IdHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(SimpleHelper(genID()))
}

func (s *Server) RootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(SimpleHelper("welcome to the fwysa api server"))
}

func (s *Server) DicewareHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	code, ok := genDiceware()
	if !ok {
		w.Write(ErrHelper("unable to generate diceware passphrase"))
	} else {
		w.Write(SimpleHelper(code))
	}
}
