package cgo

import "sync"

// MessageRecorder can be utilized to record non-SQL responses
// from the server.
//
// See the example examples/cgo/recorder on how to utilize the
// MessageRecorder.
type MessageRecorder struct {
	Messages []string
	lock     sync.Mutex
}

// NewMessageRecorder returns an initialized
// MessageRecorder.
func NewMessageRecorder() *MessageRecorder {
	return new(MessageRecorder)
}

// Reset prepares the MessageRecorder to record a new message.
func (rec *MessageRecorder) Reset() {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	rec.Messages = make([]string, 0)
}

// HandleMessage implements the MessageHandler interface.
func (rec *MessageRecorder) HandleMessage(msg Message) {
	rec.lock.Lock()
	defer rec.lock.Unlock()

	rec.Messages = append(rec.Messages, msg.Content())
}

// Text returns the received messages.
func (rec MessageRecorder) Text() []string {
	return rec.Messages
}
