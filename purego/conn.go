// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package purego

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/SAP/go-ase/libase/libdsn"
	"github.com/SAP/go-ase/libase/tds"
)

var (
	_ driver.Conn               = (*Conn)(nil)
	_ driver.ConnBeginTx        = (*Conn)(nil)
	_ driver.ConnPrepareContext = (*Conn)(nil)
	_ driver.ExecerContext      = (*Conn)(nil)
	_ driver.QueryerContext     = (*Conn)(nil)
	_ driver.Pinger             = (*Conn)(nil)
)

type Conn struct {
	Conn    *tds.Conn
	Channel *tds.Channel
	DSN     *libdsn.Info

	// TODO I don't particularly like locking statements like this
	stmts map[int]*Stmt
	// TODO: iirc conns aren't used in multiple threads at the same time
	stmtLock *sync.RWMutex
}

func NewConn(ctx context.Context, dsn *libdsn.Info) (*Conn, error) {
	return NewConnWithHooks(ctx, dsn, nil)
}

func NewConnWithHooks(ctx context.Context, dsn *libdsn.Info, envChangeHooks []tds.EnvChangeHook) (*Conn, error) {
	conn := &Conn{
		stmts:    map[int]*Stmt{},
		stmtLock: &sync.RWMutex{},
	}

	var err error
	conn.Conn, err = tds.NewConn(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("go-ase: error opening connection to TDS server: %w", err)
	}

	conn.Channel, err = conn.Conn.NewChannel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("go-ase: error opening logical channel: %w", err)
	}

	if drv.envChangeHooks != nil {
		err := conn.Channel.RegisterEnvChangeHooks(drv.envChangeHooks...)
		if err != nil {
			return nil, fmt.Errorf("go-ase: error registering driver EnvChangeHooks: %w", err)
		}
	}

	if envChangeHooks != nil {
		err := conn.Channel.RegisterEnvChangeHooks(envChangeHooks...)
		if err != nil {
			return nil, fmt.Errorf("go-ase: error registering argument EnvChangeHooks: %w", err)
		}
	}

	loginConfig, err := tds.NewLoginConfig(dsn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("go-ase: error creating login config: %w", err)
	}

	loginConfig.AppName = dsn.PropDefault("appname", "github.com/SAP/go-ase/purego")

	err = conn.Channel.Login(ctx, loginConfig)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("go-ase: error logging in: %w", err)
	}

	// TODO can this be passed another way?
	if dsn.Database != "" {
		_, err = conn.ExecContext(ctx, "use "+dsn.Database, nil)
		if err != nil {
			return nil, fmt.Errorf("go-ase: error switching to database %s: %w", dsn.Database, err)
		}
	}

	return conn, nil
}

func (c *Conn) Close() error {
	err := c.Conn.Close()
	if err != nil {
		return fmt.Errorf("go-ase: error closing TDS connection: %w", err)
	}

	return nil
}

func (c Conn) Begin() (driver.Tx, error) {
	readOnly, err := strconv.ParseBool(c.DSN.Prop("read-only"))
	if err != nil {
		return nil, fmt.Errorf("go-ase: error parsing connection property 'read-only': %w", err)
	}

	return c.BeginTx(
		context.Background(),
		driver.TxOptions{Isolation: driver.IsolationLevel(sql.LevelDefault), ReadOnly: readOnly},
	)
}

func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return nil, errors.New("go-ase: BeginTx not implemented")
}

func (c *Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if len(args) > 0 {
		stmt, err := c.NewStmt(ctx, "", query, true)
		if err != nil {
			return nil, fmt.Errorf("go-ase: error preparing statement: %w", err)
		}

		for i := range args {
			err := stmt.CheckNamedValue(&args[i])
			if err != nil {
				return nil, fmt.Errorf("go-ase: error checking argument: %w", err)
			}
		}

		result, err := stmt.ExecContext(ctx, args)
		if err != nil {
			return nil, fmt.Errorf("go-ase: error executing statement: %w", err)
		}

		return result, nil
	}

	_, result, err := c.language(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("go-ase: error executing statement: %w", err)
	}
	return result, nil
}

func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if len(args) > 0 {
		stmt, err := c.NewStmt(ctx, "", query, true)
		if err != nil {
			return nil, fmt.Errorf("go-ase: error preparing statement: %w", err)
		}

		for i := range args {
			err := stmt.CheckNamedValue(&args[i])
			if err != nil {
				return nil, fmt.Errorf("go-ase: error checking argument: %w", err)
			}
		}

		rows, err := stmt.QueryContext(ctx, args)
		if err != nil {
			return nil, fmt.Errorf("go-ase: error executing statement: %w", err)
		}

		return rows, nil
	}

	rows, _, err := c.language(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("go-ase: error executing statement: %w", err)
	}
	return rows, nil
}

func (c Conn) Ping(ctx context.Context) error {
	// TODO check rows
	// TODO implement ErrBadConn check
	_, _, err := c.language(ctx, "select 'ping'")
	if err != nil {
		return fmt.Errorf("go-ase: error pinging database: %w", err)
	}

	return nil
}
