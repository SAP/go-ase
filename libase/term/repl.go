package term

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
)

var (
	rl                 *readline.Instance
	PromptDatabaseName string
	promptMultiline    bool
)

func UpdatePrompt() {
	prompt := "> "

	if promptMultiline {
		prompt = ">>> "
	}

	if PromptDatabaseName != "" {
		prompt = PromptDatabaseName + prompt
	}

	if rl != nil {
		rl.SetPrompt(prompt)
	}
}

func Repl(db *sql.DB) error {
	var err error
	rl, err = readline.New("")
	if err != nil {
		return fmt.Errorf("term: failed to initialize readline: %w", err)
	}
	defer rl.Close()

	cmds := []string{}
	for {
		UpdatePrompt()
		line, err := rl.Readline()
		if err != nil {
			if errors.Is(err, io.EOF) {
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
			promptMultiline = true
			continue
		}

		// command is finished, reset and execute
		promptMultiline = false

		line = strings.Join(cmds, " ")
		cmds = []string{}

		err = ParseAndExecQueries(db, line)
		if err != nil {
			log.Println(err)
		}
	}
}
