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
