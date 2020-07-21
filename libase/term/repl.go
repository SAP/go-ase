package term

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
)

var rl *readline.Instance

func Repl(db *sql.DB) error {
	var err error
	rl, err = readline.New("> ")
	if err != nil {
		return fmt.Errorf("term: failed to initialize readline: %w", err)
	}
	defer rl.Close()

	cmds := []string{}
	for {
		line, err := rl.Readline()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("term: received error from readline: %w", err)
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

		err = ParseAndExecQueries(db, line)
		if err != nil {
			log.Println(err)
		}
	}
}
