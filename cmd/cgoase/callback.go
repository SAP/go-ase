package main

import (
	"log"

	"github.com/SAP/go-ase/cgo"
)

func handleMessage(msg cgo.Message) {
	if msg.MessageSeverity() == 10 {
		return
	}

	log.Println(msg.Content())
}
