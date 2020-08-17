// SPDX-FileCopyrightText: 2020 SAP SE
//
// SPDX-License-Identifier: Apache-2.0

package cgo

// #include "ctlib.h"
import "C"

var (
	// GlobalServerMessageBroker sends server messages to all registered
	// handlers.
	GlobalServerMessageBroker = newMessageBroker()
	// GlobalClientMessageBroker sends client messages to all registered
	// handlers.
	GlobalClientMessageBroker = newMessageBroker()
)

// MessageHandler describes the signature of a handler for server- and
// client messages.
type MessageHandler func(Message)

type messageBroker struct {
	handlers []MessageHandler
}

func newMessageBroker() *messageBroker {
	return new(messageBroker)
}

// RegisterHandler registers a handler for messages.
func (broker *messageBroker) RegisterHandler(handler MessageHandler) {
	broker.handlers = append(broker.handlers, handler)
}

func (broker messageBroker) recvServerMessage(csMsg *C.CS_SERVERMSG) {
	msg := newServerMessage(csMsg)

	for _, handler := range broker.handlers {
		handler(*msg)
	}
}

func (broker messageBroker) recvClientMessage(csMsg *C.CS_CLIENTMSG) {
	msg := newClientMessage(csMsg)

	for _, handler := range broker.handlers {
		handler(*msg)
	}
}
