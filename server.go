package main

import (
	"context"
	_ "github.com/go-kivik/couchdb"
	"github.com/go-kivik/kivik"
	"log"
	"net/http"
)

type Server struct {
	m           *http.ServeMux
	db          *kivik.DB
	listenAddr  string
	couchdbUrl  string
	couchdbName string
}

func NewServer(la, cdb, dbn string) *Server {
	a := Server{}
	a.listenAddr = la
	a.couchdbUrl = cdb
	a.couchdbName = dbn

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

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(SimpleHelper("OK"))
}

// Helper Handlers
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