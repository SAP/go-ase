package tds

import "bytes"

func writeString(buf *bytes.Buffer, s string, padTo int) error {
	buf.WriteString(s)
	buf.WriteString(string(make([]rune, padTo-len(s))))
	buf.WriteByte(byte(len(s)))
	return nil
}
