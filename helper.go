package main

import (
	"crypto/sha512"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/sethvargo/go-diceware/diceware"
	"golang.org/x/crypto/pbkdf2"
	"log"
	"strings"
	"time"
)

const SALT string = "fbb1fc80-4f40-4ebf-94a0-9e12aabe2746"
const ITERATIONS int = 64
const KEYLENGTH int = 64

func hash(pass string) string {
	dk := pbkdf2.Key([]byte(pass), []byte(SALT), ITERATIONS, KEYLENGTH, sha512.New)
	return base64.StdEncoding.EncodeToString(dk)
}

func genID() string {
	u := uuid.New()
	return u.String()
}

func genTimestamp() string {
	t := time.Now().UTC()
	return t.Format(time.RFC3339)
}

func genDiceware() (string, bool) {
	words, err := diceware.Generate(3)
	if err != nil {
		log.Fatal("Unable to generate diceware phrase", err)
		return "", false
	} else {
		return strings.Join(words, " "), true
	}
}
