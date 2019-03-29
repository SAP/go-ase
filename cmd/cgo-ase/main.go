package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"strings"

	_ "github.com/SAP/go-ase/cgo"
	libdsn "github.com/SAP/go-ase/libase/dsn"
	"github.com/SAP/go-ase/libase/flagslice"
	"github.com/bgentry/speakeasy"
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

func exec(db *sql.DB, q string) error {
	log.Printf("Exec '%s'", q)
	result, err := db.Exec(q)
	if err != nil {
		return fmt.Errorf("Executing the statement failed: %v", err)
	}

	return processResult(result)
}

func query(db *sql.DB, q string) error {
	log.Printf("Query '%s'", q)
	rows, err := db.Query(q)
	if err != nil {
		return fmt.Errorf("Query failed: %v", err)
	}
	defer rows.Close()

	return processRows(rows)
}

func statement(db *sql.DB, isQuery bool, query string, args []string) error {
	log.Printf("Prepare '%s'", query)
	stmt, err := db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, arg := range args {
		argSliceS := strings.Split(arg, " ")
		argSlice := make([]interface{}, len(argSliceS))
		for i, ent := range argSliceS {
			argSlice[i] = ent
		}

		if isQuery {
			log.Printf("Query prepared with '%s'", arg)
			rows, err := stmt.Query(argSlice...)
			if err != nil {
				return err
			}
			err = processRows(rows)
			rows.Close()
			if err != nil {
				return err
			}
		} else {
			log.Printf("Exec prepared with '%s'", arg)
			result, err := stmt.Exec(argSlice...)
			if err != nil {
				return err
			}
			err = processResult(result)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func subcmd(db *sql.DB, part string) error {
	partS := strings.Split(part, " ")
	cmd := partS[0]
	q := strings.Join(partS[1:], " ")

	switch cmd {
	case "exec":
		err := exec(db, q)
		if err != nil {
			return fmt.Errorf("Exec errored: %v", err)
		}
	case "query":
		err := query(db, q)
		if err != nil {
			return fmt.Errorf("Query errored: %v", err)
		}
	case "stmt":
		// cgo-ase stmt query "select * from ? where ? = ?" - "TST.dbo.test" "a" "1" - "TST.dbo.test" "a" "2"
		partS = strings.Split(q, " - ")
		queryS := strings.Split(partS[0], " ")
		isQueryS := queryS[0]
		query := strings.Join(queryS[1:], " ")

		isQuery := true
		if isQueryS == "exec" {
			isQuery = false
		}

		err := statement(db, isQuery, query, partS[1:])
		if err != nil {
			return fmt.Errorf("Statement errored: %v", err)
		}
	default:
		log.Printf("Unknown command: %s", cmd)
	}

	return nil
}

func main() {
	flag.Var(fOpts, "o", "Connection properties")
	flag.Parse()

	pass := *fPass
	var err error
	if len(pass) == 0 && len(*fUserstorekey) == 0 {
		pass, err = speakeasy.Ask("Please enter the password of user " + *fUser + ": ")
		if err != nil {
			log.Println(err)
			return
		}
	}

	dsn := libdsn.NewDsnInfo()
	dsn.Host = *fHost
	dsn.Port = *fPort
	dsn.Username = *fUser
	dsn.Password = pass
	dsn.Userstorekey = *fUserstorekey
	dsn.Database = *fDatabase

	for _, fOpt := range fOpts.Slice() {
		split := strings.SplitN(fOpt, "=", 2)
		opt := split[0]
		value := ""
		if len(split) > 1 {
			value = split[1]
		}

		dsn.ConnectProps.Set(opt, value)
	}

	db, err := sql.Open("ase", dsn.AsSimple())
	if err != nil {
		log.Printf("Error opening database connection: %v", err)
		return
	}
	defer db.Close()

	// test the database connection
	err = db.Ping()
	if err != nil {
		log.Printf("Pinging the server failed: %v", err)
		return
	}

	if len(flag.Args()) == 0 {
		return
	}

	subcmds := strings.Split(strings.Join(flag.Args(), " "), "--")
	for _, s := range subcmds {
		err = subcmd(db, strings.TrimSpace(s))
		if err != nil {
			log.Printf("Execution of '%s' resulted in error: %v", s, err)
			return
		}
	}
}
