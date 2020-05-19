package libase

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"
)

type captureConn struct {
	conn                  net.Conn
	writes, reads, merged *os.File
	writesMW, readsMW     io.Writer
}

func CaptureConn(conn net.Conn, fileprefix string) (io.ReadWriteCloser, error) {
	cc := &captureConn{conn: conn}

	if fileprefix != "" {
		fileprefix += "-"
	}

	filenameTemplate := "/tmp/%scapture-%s-%s.bin"
	// TODO
	timeStamp := time.Now().Format(time.RFC3339)

	writes, err := os.OpenFile(fmt.Sprintf(filenameTemplate, fileprefix, timeStamp, "input"), os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return nil, fmt.Errorf("error opening input capture file: %w", err)
	}
	cc.writes = writes

	reads, err := os.OpenFile(fmt.Sprintf(filenameTemplate, fileprefix, timeStamp, "output"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("error opening output capture file: %w", err)
	}
	cc.reads = reads

	merged, err := os.OpenFile(fmt.Sprintf(filenameTemplate, fileprefix, timeStamp, "merged"), os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, fmt.Errorf("error opening merged capture file: %w", err)
	}
	cc.merged = merged

	cc.writesMW = io.MultiWriter(conn, writes, merged)
	cc.readsMW = io.MultiWriter(reads, merged)

	return cc, nil
}

func (cc *captureConn) Write(bs []byte) (int, error) {
	return cc.writesMW.Write(bs)
}

func (cc *captureConn) Read(bs []byte) (int, error) {
	n, err := cc.conn.Read(bs)

	_, errCopy := cc.readsMW.Write(bs)
	if errCopy != nil {
		return n, fmt.Errorf("failed to write to capture files: %w", err)
	}

	return n, err
}

func closeClosers(closers map[string]io.Closer) error {
	for name, closer := range closers {
		err := closer.Close()
		if err != nil {
			return fmt.Errorf("failed to close %s: %w", name, err)
		}
	}
	return nil
}

func (cc *captureConn) Close() error {
	return closeClosers(map[string]io.Closer{
		"write capture file":  cc.writes,
		"reads capture file":  cc.reads,
		"merged capture file": cc.merged,
		"connection":          cc.conn,
	})
}
