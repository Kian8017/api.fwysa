package main

import (
	"encoding/json"
	"log"
)

type Simple struct {
	Ok     bool   `json:"ok"`
	Result string `json:"result"`
}

func (s Simple) JSON() []byte {
	m, err := json.Marshal(s)
	if err != nil {
		log.Fatal("Unable to stringify simpleStruct", err)
	}
	return m
}

func NewSimple(o bool, r string) Simple {
	return Simple{Ok: o, Result: r}
}

func ErrHelper(r string) []byte {
	return NewSimple(false, r).JSON()
}

func SimpleHelper(r string) []byte {
	return NewSimple(true, r).JSON()
}

type AuthDocument struct {
	Id       string `json:"_id"`
	Rev      string `json:"_rev"`
	Type     string `json:"type"`
	UserID   string `json:"userID"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Password string `json:"password"`
}

type LoginDetails struct {
	Ok       bool   `json:"ok"`
	Role     string `json:"role"`
	UserID   string `json:"userID"`
	DBString string `json:"couch"`
}

func (l LoginDetails) JSON() []byte {
	m, err := json.Marshal(l)
	if err != nil {
		log.Fatal("Unable to stringify loginDetails", err)
	}
	return m
}

func NewLoginDetails(r, uid, dbs string) LoginDetails {
	return LoginDetails{Ok: true, UserID: uid, DBString: dbs}
}

func LoginHelper(r, uid, dbs string) []byte {
	return NewLoginDetails(r, uid, dbs).JSON()
}
