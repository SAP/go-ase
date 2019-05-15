package cgo

// #include "ctlib.h"
import "C"

var (
	// GlobalServerMessageBroker sends server messages to all registered
	// handlers.
	GlobalServerMessageBroker = newServerMessageBroker()
	// GlobalClientMessageBroker sends client messages to all registered
	// handlers.
	GlobalClientMessageBroker = newClientMessageBroker()
)

// ServerMessageHandler describes the signature of a handler for server
// messages.
type ServerMessageHandler func(ServerMessage)

// ClientMessageHandler describes the signature of a handler for client
// messages.
type ClientMessageHandler func(ClientMessage)

type serverMessageBroker struct {
	handlers []ServerMessageHandler
}

func newServerMessageBroker() *serverMessageBroker {
	return new(serverMessageBroker)
}

type clientMessageBroker struct {
	handlers []ClientMessageHandler
}

func newClientMessageBroker() *clientMessageBroker {
	return new(clientMessageBroker)
}

// RegisterHandler registers a handler for server messages.
func (broker *serverMessageBroker) RegisterHandler(handler ServerMessageHandler) {
	broker.handlers = append(broker.handlers, handler)
}

// RegisterHandler registers a handler for client messages.
func (broker *clientMessageBroker) RegisterHandler(handler ClientMessageHandler) {
	broker.handlers = append(broker.handlers, handler)
}

func (broker serverMessageBroker) recvMessage(csMsg *C.CS_SERVERMSG) {
	msg := newServerMessage(csMsg)

	for _, handler := range broker.handlers {
		handler(*msg)
	}
}

func (broker clientMessageBroker) recvMessage(csMsg *C.CS_CLIENTMSG) {
	msg := newClientMessage(csMsg)

	for _, handler := range broker.handlers {
		handler(*msg)
	}
}
