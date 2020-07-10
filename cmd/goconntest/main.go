package main

import (
	"fmt"
	"log"

	"github.com/SAP/go-ase/libase/libdsn"
	"github.com/SAP/go-ase/libase/tds"
)

func main() {
	log.Printf("Retrieving DSNInfo")
	dsn := libdsn.NewDsnInfoFromEnv("")

	log.Printf("Dialing to server")
	c, err := tds.Dial("tcp",
		fmt.Sprintf("%s:%s", dsn.Host, dsn.Port),
	)
	if err != nil {
		log.Printf("Failed to dial server: %v", err)
		return
	}
	defer func() {
		err := c.Close()
		if err != nil {
			log.Printf("Error closing connection: %v", err)
		}
	}()

	log.Printf("Preparing login packet")
	conf, err := tds.NewLoginConfig(dsn)
	if err != nil {
		log.Printf("Failed to prepare login packet: %v", err)
		return
	}
	conf.AppName = "goconntest"

	log.Printf("Logging in")
	err = c.Login(conf)
	if err != nil {
		log.Printf("Login failed: %v", err)
		return
	}

	msg := tds.NewMessage()
	langPkg := &tds.LanguagePackage{
		Status: tds.TDS_LANGUAGE_NOARGS,
		Cmd:    "sp_help",
	}
	msg.AddPackage(langPkg)

	err = c.Send(*msg)
	if err != nil {
		log.Printf("Sending language command failed: %v", err)
		return
	}

	_, err = c.Receive()
	if err != nil {
		log.Printf("Receiving response to language command failed: %v", err)
		return
	}
}
