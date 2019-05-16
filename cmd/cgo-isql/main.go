package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	ase "github.com/SAP/go-ase/cgo"
	libdsn "github.com/SAP/go-ase/libase/dsn"
	"github.com/SAP/go-ase/libase/flagslice"
)

var (
	fHost         = flag.String("H", "", "database hostname")
	fPort         = flag.String("P", "", "database sql port")
	fUser         = flag.String("u", "", "database user name")
	fPass         = flag.String("p", "", "database user password")
	fUserstorekey = flag.String("k", "", "userstorekey")
	fDatabase     = flag.String("D", "", "database")

	fOpts = &flagslice.FlagStringSlice{}
)

func openDB() (*ase.Connection, error) {
	dsn := libdsn.NewDsnInfoFromEnv("")

	if *fHost != "" {
		dsn.Host = *fHost
	}

	if *fPort != "" {
		dsn.Port = *fPort
	}

	if *fUser != "" {
		dsn.Username = *fUser
	}

	if *fPass != "" {
		dsn.Password = *fPass
	}

	if *fUserstorekey != "" {
		dsn.Userstorekey = *fUserstorekey
	}

	if *fDatabase != "" {
		dsn.Database = *fDatabase
	}

	for _, fOpt := range fOpts.Slice() {
		split := strings.SplitN(fOpt, "=", 2)
		opt := split[0]
		value := ""
		if len(split) > 1 {
			value = split[1]
		}

		dsn.ConnectProps.Set(opt, value)
	}

	conn, err := ase.NewConnection(nil, *dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to open connection: %v", err)
	}

	return conn, nil
}

func serverMessagePrint(msg ase.Message) {
	fmt.Fprintf(os.Stderr, "\r%s", msg.Content())

	if rl != nil {
		rl.Refresh()
	}
}

func main() {
	flag.Var(fOpts, "o", "Connection properties")
	flag.Parse()

	ase.GlobalServerMessageBroker.RegisterHandler(serverMessagePrint)
	ase.GlobalClientMessageBroker.RegisterHandler(serverMessagePrint)

	conn, err := openDB()
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return
	}
	defer conn.Close()

	err = repl(conn)
	if err != nil {
		log.Println(err)
		return
	}
}
