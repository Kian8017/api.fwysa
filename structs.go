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
	Rev      string `json:"_rev,omitempty"`
	Type     string `json:"type"`
	Created  string `json:"created"`
	UserID   string `json:"userID"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func NewAuthDocument(uid, un, r, p string) AuthDocument {
	return AuthDocument{
		Id:       genID(),
		Type:     "auth",
		Created:  genTimestamp(),
		UserID:   uid,
		Username: un,
		Role:     r,
		Password: hash(p),
	}
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
	return LoginDetails{Ok: true, Role: r, UserID: uid, DBString: dbs}
}

func LoginHelper(r, uid, dbs string) []byte {
	return NewLoginDetails(r, uid, dbs).JSON()
}

type PendingAuthDocument struct {
	Id      string `json:"_id"`
	Rev     string `json:"_rev,omitempty"`
	Type    string `json:"type"`
	Created string `json:"created"`
	UserID  string `json:"userID"`
	Code    string `json:"code"`
	Role    string `json:"role"`
}

func NewPendingAuthDocument(uid, role string) PendingAuthDocument {
	return PendingAuthDocument{
		Id:      genID(),
		Type:    "pendingauth",
		Created: genTimestamp(),
		UserID:  uid,
		Code:    genDiceware(),
		Role:    role,
	}
}

type PageSection struct {
	Parent      int    `json:"parent"`
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Expanded    bool   `json:"expanded"`
	Highlighted bool   `json:"highlighted"`
	Contents    string `json:"contents"`
}
