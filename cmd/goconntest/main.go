package main

import (
	"context"
	"fmt"
	"log"

	"github.com/SAP/go-ase/libase/libdsn"
	"github.com/SAP/go-ase/libase/tds"
)

func main() {
	log.Printf("Retrieving DSNInfo")
	dsn := libdsn.NewDsnInfoFromEnv("")

	log.Printf("Dialing to server")
	tdsConn, err := tds.NewTDSConn(
		context.Background(),
		"tcp",
		fmt.Sprintf("%s:%s", dsn.Host, dsn.Port),
	)
	if err != nil {
		log.Printf("Failed to dial server: %v", err)
		return
	}
	defer func() {
		err := tdsConn.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	// Channel 0
	tdsCh0, err := tdsConn.NewTDSChannel(100)
	if err != nil {
		log.Printf("Failed to open channel: %w", err)
		return
	}
	defer tdsCh0.Close()

	log.Printf("Preparing login packet")
	conf, err := tds.NewLoginConfig(dsn)
	if err != nil {
		log.Printf("Failed to prepare login packet: %v", err)
		return
	}
	conf.AppName = "goconntest"

	log.Printf("Logging in")
	err = tdsCh0.Login(conf)
	if err != nil {
		log.Printf("Login failed: %v", err)
		return
	}

	langPkg := &tds.LanguagePackage{
		Status: tds.TDS_LANGUAGE_NOARGS,
		Cmd:    "sp_help",
	}

	log.Printf("Sending language command")
	err = tdsCh0.QueuePackage(langPkg)
	if err != nil {
		log.Printf("Error sending package: %w", err)
		return
	}

	err = tdsCh0.SendRemainingPackets()
	if err != nil {
		log.Printf("Error sending packets: %w", err)
		return
	}

	pkg, err := tdsCh0.NextPackage(true)
	if err != nil {
		log.Printf("Error receiving package: %w", err)
		return
	}

	log.Printf("%#v", pkg)
}
