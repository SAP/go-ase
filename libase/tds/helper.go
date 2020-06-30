package tds

import "io"

// writeString writes s padded to padTo and its length to buf.
func writeString(stream io.Writer, s string, padTo int) error {
	stream.Write([]byte(s))
	stream.Write(make([]byte, padTo-len(s)))
	stream.Write([]byte{byte(len(s))})
	return nil
}
