// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package ase

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	"github.com/SAP/go-dblib"
	"github.com/SAP/go-dblib/tds"
)

var (
	_ driver.ConnBeginTx = (*Conn)(nil)
	_ driver.Tx          = (*Transaction)(nil)
)

func DefaultTxOptions() driver.TxOptions {
	return driver.TxOptions{
		Isolation: driver.IsolationLevel(sql.LevelDefault),
		ReadOnly:  false,
	}
}

type Transaction struct {
	conn *Conn
	name string
}

func (tx Transaction) Name() string {
	return tx.name
}

func (c *Conn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), DefaultTxOptions())
}

func (c *Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return c.NewTransaction(ctx, opts, "")
}

func (c *Conn) NewTransaction(ctx context.Context, opts driver.TxOptions, name string) (*Transaction, error) {
	tx := &Transaction{
		conn: c,
		name: name,
	}

	return tx, tx.begin(ctx, opts)
}

func (tx Transaction) begin(ctx context.Context, opts driver.TxOptions) error {
	if opts.ReadOnly {
		return errors.New("go-ase: ASE does not support read-only transactions")
	}

	isolationLvl, err := dblib.ASEIsolationLevelFromGo(sql.IsolationLevel(opts.Isolation))
	if err != nil {
		return fmt.Errorf("go-ase: error mapping sql.IsolationLevel to ASE isolation level: %w", err)
	}

	if isolationLvl == dblib.ASELevelInvalid {
		return fmt.Errorf("go-ase: sql.IsolationLevel %s has no equivalent ASE isolation level", sql.IsolationLevel(opts.Isolation))
	}

	if _, _, err := tx.conn.GenericExec(ctx, "begin transaction "+tx.name, nil); err != nil {
		return fmt.Errorf("go-ase: error initializing transaction: %w", err)
	}

	optIsolationPkg := &tds.OptionCmdPackage{
		Cmd:       tds.TDS_OPT_SET,
		Option:    tds.TDS_OPT_ISOLATION,
		OptionArg: []byte{byte(isolationLvl)},
	}
	if err := tx.conn.Channel.QueuePackage(ctx, optIsolationPkg); err != nil {
		return fmt.Errorf("go-ase: error queueing package: %w", err)
	}

	return nil
}

func (tx Transaction) NewTransaction(ctx context.Context, opts driver.TxOptions) (*Transaction, error) {
	newTx := &Transaction{
		conn: tx.conn,
	}

	return newTx, newTx.begin(ctx, opts)
}

func (tx Transaction) Commit() error {
	if _, _, err := tx.conn.GenericExec(context.Background(), "commit "+tx.name, nil); err != nil {
		return fmt.Errorf("go-ase: error committing transaction: %w", err)
	}
	return nil
}

func (tx Transaction) Rollback() error {
	if _, _, err := tx.conn.GenericExec(context.Background(), "rollback "+tx.name, nil); err != nil {
		return fmt.Errorf("go-ase: error rolling back transaction: %w", err)
	}
	return nil
}
