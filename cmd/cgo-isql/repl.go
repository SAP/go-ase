package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/SAP/go-ase/cgo"
	"github.com/chzyer/readline"
)

var rl *readline.Instance

func repl(conn *cgo.Connection) error {
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
		cmds = []string{}
		rl.SetPrompt("> ")

		err = parseAndExecQueries(conn, line)
		if err != nil {
			log.Println(err)
		}
	}
}

func parseAndExecQueries(conn *cgo.Connection, line string) error {
	builder := strings.Builder{}
	currentlyQuoted := false

	for _, chr := range line {
		switch chr {
		case '"', '\'':
			if currentlyQuoted {
				currentlyQuoted = false
				builder.WriteRune(chr)
			} else {
				currentlyQuoted = true
				builder.WriteRune(chr)
			}
		case ';':
			if currentlyQuoted {
				builder.WriteRune(chr)
			} else {
				err := process(conn, builder.String())
				if err != nil {
					return fmt.Errorf("Failed to process query: %v", err)
				}
				builder.Reset()
			}
		default:
			builder.WriteRune(chr)
		}
	}

	return nil
}
