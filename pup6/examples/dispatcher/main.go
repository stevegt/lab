package main

import (
	"github.com/multiformats/go-multihash"
	. "github.com/stevegt/goadapt"
)

// Message is a parsed PUP-pattern message.
type Message struct {
	// Id is the multihash prefix of the message.
	Id      multihash.Multihash
	Payload []byte
}

// Handler is an interface that must be implemented by all handlers
// that register with the dispatcher.
type Handler interface {
	// Claim returns true if the handler claims the message.
	Claim(msg *Message) bool
	HandleMessage(msg *Message, responses chan *Message)
}

// Dispatcher is a PUP dispatcher that routes messages and streams to
// the appropriate handler.
type Dispatcher struct {
	// Handlers is a list of handlers that are registered with the
	// dispatcher.  Each incoming message is routed to all handlers
	// that claim to be able to process it.
	Handlers []Handler
}

// HandleMessage handles a message by routing it to all handlers that
// claim it. Handlers return errors and responses in the responses
// channel.
func (d *Dispatcher) HandleMessage(msg *Message, responses chan *Message) {
	claimed := false
	for _, h := range d.Handlers {
		if h.Claim(msg) {
			claimed = true
			go h.HandleMessage(msg, responses)
		}
	}
	if !claimed {
		responses <- NewErrorMessage(msg, "No handler claimed the message.")
	}
}

// RegisterHandler registers a handler with the dispatcher.
func (d *Dispatcher) RegisterHandler(h Handler) {
	d.Handlers = append(d.Handlers, h)
}

// NewDispatcher creates a new dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{}
}

// EchoHandler is a handler that echos messages back to the sender.
type EchoHandler struct{}

// Claim returns true if the handler claims the message.
func (h *EchoHandler) Claim(msg *Message) bool {

	return msg.Type == "echo"
}

func main() {
	// Create a new dispatcher.
	dispatcher := NewDispatcher()

	// Register a handler that handles the "echo" message.
	dispatcher.RegisterHandler(&EchoHandler{})

	// receive messages from stdin and send responses to stdout.
	dispatcher.Run()

	Pl("vim-go")

}
