package main

import (
	"context"
	"log"

	"github.com/SAP/go-ase/libase/libdsn"
	"github.com/SAP/go-ase/libase/tds"
)

func main() {
	log.Printf("Retrieving DSNInfo")
	dsn := libdsn.NewDsnInfoFromEnv("")

	log.Printf("Dialing to server")
	conn, err := tds.NewConn(context.Background(), dsn)
	if err != nil {
		log.Printf("Failed to dial server: %v", err)
		return
	}
	defer func() {
		err := conn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	// Channel 0
	ch0, err := conn.NewChannel()
	if err != nil {
		log.Printf("Failed to open channel: %w", err)
		return
	}

	log.Printf("Preparing login packet")
	conf, err := tds.NewLoginConfig(dsn)
	if err != nil {
		log.Printf("Failed to prepare login packet: %v", err)
		return
	}
	conf.AppName = "goconntest"

	log.Printf("Logging in")
	err = ch0.Login(context.Background(), conf)
	if err != nil {
		log.Printf("Login failed: %v", err)
		return
	}

	langPkg := &tds.LanguagePackage{
		Status: tds.TDS_LANGUAGE_NOARGS,
		Cmd:    "select 'ping'",
	}

	log.Printf("Sending language command")
	err = ch0.SendPackage(context.Background(), langPkg)
	if err != nil {
		log.Printf("Error sending package: %w", err)
		return
	}

	// rowfmt
	_, err = ch0.NextPackage(true)
	if err != nil {
		log.Printf("Error receiving package: %w", err)
		return
	}

	// row
	_, err = ch0.NextPackage(true)
	if err != nil {
		log.Printf("Error receiving package: %w", err)
		return
	}

	// done
	_, err = ch0.NextPackage(true)
	if err != nil {
		log.Printf("Error receiving package: %w", err)
		return
	}
}
