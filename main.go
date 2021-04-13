package main

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	couchdbUrl := os.Getenv("COUCHDB_URL")
	couchdbName := os.Getenv("COUCHDB_NAME")
	listenAddr := os.Getenv("LISTEN_ADDR")

	// actions:
	//   login
	//   genpendingauth
	//   createauth

	s := NewServer(listenAddr, couchdbUrl, couchdbName)

	s.Run()
}
