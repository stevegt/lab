package stategraph

import (
	"fmt"

	"github.com/multiformats/go-multihash"
	. "github.com/stevegt/goadapt"
)

/*
   * algorithms:
       * runtime responds to a CAS read call: cas.read(state, hash)
           * if content not local:
               * send peers a ‘read’ message with runtime’s public key
               * return promise to caller
           * execute transition: exec(state, readfunc, hash)
               * transition call includes consensus and security checks; abort if fail
           * return newstate, content to caller
       * runtime responds to a CAS write call: cas.write(data)
           * add origin and random nonce, hash, sign hash, and store in local CAS
           * return the hash to the caller
           * do not create a new state -- wait for a later read to reference the chunk
       * runtime responds to a transition (exec) call:  exec(state, func, args…)
           * execute consensus test, abort if fail
               * test(state, func, args…)
               * consensus includes security checks
           * execute transition function in sandbox
               * makes read and write CAS calls
           * hash new state and add it to the state graph
           * sign and publish a message with the old and new state hashes, transition hash, hash of arg set
       * runtime processes an incoming message:
           * if transition:
               * call exec(state, func, args…)
           * if read request:
               * call read(state, hash)
               * encrypt the response with the requestor’s public key
                   * requestor might be different (upstream) from peer
               * send response to peer
           * if read response:
               * decrypt response with the runtime’s private key
               * fulfill promise
       * runtime performs GC:
           * prune state graph
               * squash or remove stale branches
           * remove stale CAS entries not referenced by state graph
   * definition of “publish”:
       * for each node in state graph, there is a set of subscribers, stored in CAS, referenced from a separate branch in the state graph
       * subscribers add or remove their own entries using CAS writes
       * when the runtime executes a transition, it sends the resulting message to the peers who are in the subscriber list for the old state
*/

// An address is a large cryptographic hash.  It can be used to
// uniquely identify any object: a node in the state graph, functions
// used in transitions, content in the CAS, journal messages, etc. We
// use the multihash standard to encode the hash, allowing us to use
// different hash algorithms interchangeably.  See the
// http://github.com/multiformats project for more information.
type Address multihash.Multihash

// Message is a message sent between peers.
// XXX message is a program
// XXX transition is a hyperedge
// XXX message either describes:
// - parent states, function, new state
// - parent states, function, nil (query)
// - parent states, function, new state, encrypted payload (reply)
type Message struct {
	// sender address
	From Address
	// recipient address
	To Address
	// signature of the message
	Signature []byte
	// Address of the starting state nodes
	States []Address
	// message type
	Opcode MessageType
	// - address of the function to execute if this is a transition
	// - address of the content to read if this is a read request
	Operand Address
	// byte array of the encrypted message if this is a reply to a read
	Payload []byte
}

// Runtime is the main object that manages the state graph.
type Runtime struct {
}

/*
OnMessage is called when a message is received from a peer.
  - if transition:
  - call exec(state, func, args…)
  - if read request:
  - call read(state, hash)
  - encrypt the response with the requestor’s public key
  - requestor might be different (upstream) from peer
  - send response to peer
  - if read response:
  - decrypt response with the runtime’s private key
  - fulfill promise
*/
func (r *Runtime) OnMessage(m Message) (err error) {
	typ := m.Type()
	switch typ {
	case "transition":
		r.exec(m.State(), m.Func(), m.Args())
	case "read":
		payload := r.read(m.State(), m.Hash())
		clearMsg := r.MkMsg("reply", m.Sender(), payload)
		cryptMsg := r.Encrypt(clearMsg)
		r.Send(cryptMsg)
	case "reply":
		clearMsg := r.Decrypt(m)
		r.Fulfill(clearMsg)
	default:
		err = fmt.Errorf("unknown message type: %s", typ)
	}
	return
}

// State is a node in the state graph.
type State struct {
}

// Transition is an edge in the state graph.
type Transition struct {
}

// CAS is the content-addressable storage for the runtime.
