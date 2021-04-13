package main

import (
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
)

func hashHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	pass, ok := query["pass"]
	if !ok {
		w.Write(ErrHelper("pass query parameter not present"))
		return
	}

	w.Write(SimpleHelper(hash(pass[0])))
}

func timestampHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(SimpleHelper(genTimestamp()))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write(SimpleHelper("welcome to the fwysa api server"))
}

func dicewareHandler(w http.ResponseWriter, r *http.Request) {
	code, ok := genDiceware()
	if !ok {
		w.Write(ErrHelper("unable to generate diceware passphrase"))
	} else {
		w.Write(SimpleHelper(code))
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// couchdbUrl := os.Getenv("COUCHDB_URL")
	listenAddr := os.Getenv("LISTEN_ADDR")

	// actions:
	//   login
	//   genpendingauth
	//   createauth

	http.HandleFunc("/hash", hashHandler)
	http.HandleFunc("/timestamp", timestampHandler)
	http.HandleFunc("/diceware", dicewareHandler)
	http.HandleFunc("/", rootHandler)

	log.Println("Server listening at", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
