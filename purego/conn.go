package purego

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"

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
	DSN     *libdsn.DsnInfo
}

func NewConn(ctx context.Context, dsn *libdsn.DsnInfo) (*Conn, error) {
	conn := &Conn{}

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

func (c Conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return nil, errors.New("go-ase: BeginTx not implemented")
}

func (c Conn) Prepare(query string) (driver.Stmt, error) {
	return c.PrepareContext(context.Background(), query)
}

func (c Conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	return nil, errors.New("go-ase: PrepareContext not implemented")
}

func (c Conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if len(args) > 0 {
		return nil, errors.New("go-ase: args not implemented")
	}

	_, result, err := c.language(ctx, query)
	return result, err
}

func (c Conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if len(args) > 0 {
		return nil, errors.New("go-ase: args not implemented")
	}

	rows, _, err := c.language(ctx, query)
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
