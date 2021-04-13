package main

import (
	"log"
	"net/http"
)

type Server struct {
	m          *http.ServeMux
	listenAddr string
}

func NewServer(la string) *Server {
	a := Server{}
	a.listenAddr = la

	sm := http.NewServeMux()
	a.m = sm
	// Example handlers
	sm.HandleFunc("/hash", a.HashHandler)
	sm.HandleFunc("/timestamp", a.TimestampHandler)
	sm.HandleFunc("/id", a.IdHandler)
	sm.HandleFunc("/diceware", a.DicewareHandler)

	// Root handler
	sm.HandleFunc("/", a.RootHandler)

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
