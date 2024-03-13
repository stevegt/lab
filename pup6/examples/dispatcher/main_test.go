package main

import (
	"testing"

	"github.com/multiformats/go-multihash"
	. "github.com/stevegt/goadapt"
)

// TestEchoHandler tests the EchoHandler.
func TestEchoHandler(t *testing.T) {
	// Create a new dispatcher.
	d := NewDispatcher()

	// Register the echo handler.
	d.RegisterHandler(&EchoHandler{})

	// Create a message.
	m := &Message{
		Id:      multihash.Multihash{},
		Payload: []byte("Hello, world!"),
	}

	// Create a channel for responses.
	responses := make(chan *Message)

	// Dispatch the message.
	d.HandleMessage(m, responses)

	// Get the response.
	r := <-responses

	// Check the response.
	if string(r.Payload) != "Hello, world!" {
		t.Error("EchoHandler failed to echo the message.")
	}
}
