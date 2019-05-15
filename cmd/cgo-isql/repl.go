package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	ase "github.com/SAP/go-ase/cgo"
	"github.com/chzyer/readline"
)

var rl *readline.Instance

func repl(conn *ase.Connection) error {
	var err error
	rl, err = readline.New("> ")
	if err != nil {
		return fmt.Errorf("Failed to initialize readline: %v", err)
	}
	defer rl.Close()

	cmds := []string{}
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("Received error from readline: %v", err)
		}

		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		cmds = append(cmds, line)

		if !strings.HasSuffix(line, ";") {
			rl.SetPrompt(">>> ")
			continue
		}

		line = strings.Join(cmds, " ")

		line = line[:len(line)-1]

		cmds = []string{}

		rl.SetPrompt("> ")

		cmd, err := conn.GenericExec(context.Background(), line)
		if err != nil {
			log.Printf("Query failed: %v", err)
			continue
		}
		defer cmd.Drop()

		for {
			rows, result, err := cmd.Response()
			if err != nil {
				if err == io.EOF {
					break
				}
				cmd.Cancel()
				log.Printf("Reading response failed: %v", err)
				break
			}

			if rows != nil {
				err = processRows(rows)
				if err != nil {
					log.Printf("Error processing rows: %v", err)
				}
			}

			if result != nil {
				err = processResult(result)
				if err != nil {
					log.Printf("error processing result: %v", err)
				}
			}
		}
	}
}
