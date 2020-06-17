package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/SAP/go-ase/cgo"
	"github.com/SAP/go-ase/libase/flagslice"
	"github.com/SAP/go-ase/libase/libdsn"
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

func openDB() (*cgo.Connection, error) {
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

	cgo.GlobalServerMessageBroker.RegisterHandler(handleMessage)
	cgo.GlobalClientMessageBroker.RegisterHandler(handleMessage)

	conn, err := cgo.NewConnection(nil, *dsn)
	if err != nil {
		return nil, fmt.Errorf("Failed to open connection: %v", err)
	}

	return conn, nil
}

func serverMessagePrint(msg cgo.Message) {
	fmt.Fprintf(os.Stderr, "\r%s", msg.Content())

	if rl != nil {
		rl.Refresh()
	}
}

func main() {
	err := doMain()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func doMain() error {
	flag.Var(fOpts, "o", "Connection properties")
	flag.Parse()

	cgo.GlobalServerMessageBroker.RegisterHandler(serverMessagePrint)
	cgo.GlobalClientMessageBroker.RegisterHandler(serverMessagePrint)

	conn, err := openDB()
	if err != nil {
		return fmt.Errorf("Failed to connect to database: %v", err)
	}
	defer conn.Close()

	if len(flag.Args()) > 0 {
		// Positional arguments were supplied, execute these as SQL
		// statements
		query := strings.Join(flag.Args(), " ") + ";"
		err = parseAndExecQueries(conn, query)
	} else {
		err = repl(conn)
	}

	return err
}
