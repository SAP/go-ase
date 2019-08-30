package tds

import "fmt"

func (tds *TDSConn) Receive() (*Message, error) {
	msg := &Message{}

	err := msg.readFrom(tds.conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read message: %v", err)
	}

	return msg, nil
}
