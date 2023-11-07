# Pattern for Universal Protocols (PUP)

The Pattern for Universal Protocols (PUP) is a design pattern for
communication between dissimilar systems, languages, eras, and
architectures.  With this pattern, you can use a single listener or
dispatcher to handle messages for multiple versions of multiple
protocols.  To use this pattern, calculate a multihash from the code
of a protocol's handler, prefix each message with the multihash, and
write your dispatcher to parse and route messages based on the
multihash prefix.

This README.md file is the PUP specification.  This repository also
contains an example implementation of a PUP implementation in the
`examples` directory.

## PUP Specification

### Framing

A PUP dispatcher operates at OSI layer 5, and doesn't specify framing
or message delimiters -- use a lower-level protocol to frame messages.
For example, you might use UDP, websocket, or MQTT for framing.

The PUP patterns assumes that the lower-level protocol is
reliable.  If you are using an unreliable protocol, you will need to
add a mechanism in a handler to detect and retransmit lost messages.

### Dispatcher

Write your dispatcher to parse the multihash prefix from each message
and route the message to the appropriate handler.  The dispatcher
should be able to handle any number of protocols, and should be able
to handle protocols that are added or removed at runtime.

You'll want to make sure that the dispatcher is capable of handling
message receipt in parallel.  Use threading or forking to handle
multiple messages at once -- avoid blocking receipt of a message while
processing another message.

### Message Format

A PUP message consists of the multihash protocol identifier followed
by a newline (0x0a) and the layer 6 payload.  The payload can be any
format or length.  

There is no content length field in the layer 5 message; the framing
layer is the only mechanism for determining the length of a message.
However, a layer 6 handler could implement a content length field in
the message payload if needed.

### Protocol Identifier

The protocol identifier is a hex-encoded multihash that uniquely
identifies the protocol.  The multihash is a hash that is prefixed
with a hash function identifier, so is itself designed to be
future-proof. See [the multihash
project](http://github.com/multiformats/multihash) for more
information.

You would typically calculate the multihash from the code of the protocol's 
handler -- this would allow runtime verification of handler code and
allow the sender to precisely specify the handler to be used.  If
using this method, you will likely want to design your layer 6
protocol to include a manifest in the message payload that acts as a
symbol table to associate handler hashes with function names, allowing 
the sender to specify function versions deeper in the call tree.  XXX
provide example

As an alternative, you could calculate the multihash from the protocol's
canonical specification document.  The document might be an RFC, a W3C
specification, or any document published by a standards body or open
source project. The document must be canonical, meaning that it is a
single document that is the definitive specification for the protocol,
and should be immutable and available from multiple sources.  If you
use a document as the basis for the multihash, you should include a
citation for the document and its hash in the source code of the handler.

