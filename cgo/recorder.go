package cgo

import "sync"

// ServerMessageRecorder can be utilized to record non-SQL responses
// from the server.
//
// See the example examples/cgo/recorder on how to utilize the
// ServerMessageRecorder.
type ServerMessageRecorder struct {
	Messages []string
	lock     sync.Mutex
}

// NewServerMessageRecorder returns an initialized
// ServerMessageRecorder.
func NewServerMessageRecorder() *ServerMessageRecorder {
	return new(ServerMessageRecorder)
}

// Resets prepares the ServerMessageRecorder to record a new message.
func (rec *ServerMessageRecorder) Reset() {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	rec.Messages = make([]string, 0)
}

// Handle implements the ServerMessageHandler interface.
func (rec *ServerMessageRecorder) Handle(msg ServerMessage) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	rec.Messages = append(rec.Messages, msg.Content())
}

// Text returns the received messages as a newline-separated string and
// an error if not all lines were received.
//
// The recorder will be reset after the method has returned.
func (rec ServerMessageRecorder) Text() (int, []string, error) {
	defer rec.Reset()
	rec.lock.Lock()
	defer rec.lock.Unlock()

	return len(rec.Messages), rec.Messages, nil
}
