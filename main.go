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

	// couchdbUrl := os.Getenv("COUCHDB_URL")
	listenAddr := os.Getenv("LISTEN_ADDR")

	// actions:
	//   login
	//   genpendingauth
	//   createauth

	s := NewServer(listenAddr)

	s.Run()
}
