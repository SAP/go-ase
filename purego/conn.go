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
		if _, err = conn.ExecContext(ctx, "use "+dsn.Database, nil); err != nil {
			return nil, fmt.Errorf("go-ase: error switching to database %s: %w", dsn.Database, err)
		}
	}

	return conn, nil
}

func (c *Conn) Close() error {
	if err := c.Conn.Close(); err != nil {
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
	_, result, err := c.GenericExec(ctx, query, args)
	return result, err
}

func (c *Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	rows, _, err := c.GenericExec(ctx, query, args)
	return rows, err
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
