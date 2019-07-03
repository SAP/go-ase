package cgotest

import (
	"log"

	"github.com/SAP/go-ase/cgo"
)

func genMessageHandler() cgo.MessageHandler {
	return func(msg cgo.Message) {
		// Ignore CS_SV_INFORM
		if msg.MessageSeverity() == 10 {
			return
		}

		log.Printf("Callback: %v", msg.Content())
	}
}
