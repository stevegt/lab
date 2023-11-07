# Pattern for Universal Protocols (PUP)

The Pattern for Universal Protocols (PUP) is a design pattern for
communication between dissimilar systems, languages, eras, and
architectures.  With this pattern, you can use a single listener or
dispatcher to handle messages and streams in multiple protocols.  To
use this pattern, you will calculate a multihash from any protocol's
canonical specification document, prefix each message or stream with
the multihash, and write your dispatcher to parse and route messages
based on the multihash prefix.

This README.md file is the PUP specification.  This repository also
contains an example implementation of a PUP dispatcher in the
`examples` directory.

## PUP Specification

### Framing

PUP doesn't specify framing or message delimiters -- use a lower-level
protocol to frame messages and streams.  For example, you might use 
UDP, websocket, or MQTT to frame messages, or TCP for streams.  

The PUP patterns assumes that the lower-level protocol is
reliable.  If you are using an unreliable protocol, you will need to
add a mechanism to detect and retransmit lost messages.

For the rest of this document, we will use the term "message" to refer
to both messages and streams; a stream is considered to be a message
of indefinite length.  

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

A PUP message consists of a protocol identifier followed by a newline
(0x0a) and the payload.  The payload can be any format or length.  

There is no content length field in the message; the framing layer is
the only mechanism for determining the length of the message.  

### Protocol Identifier

The protocol identifier is a hex-encoded multihash that uniquely
identifies the protocol.  The multihash is a hash that is prefixed
with a hash function identifier, so is itself designed to be
future-proof. See [the multihash
project](http://github.com/multiformats/multihash) for more
information.

You would typically calculate the multihash from the protocol's
canonical specification document.  The document might be an RFC, a W3C
specification, or any document published by a standards body or open
source project. The document must be canonical, meaning that it is a
single document that is the definitive specification for the protocol,
and should be immutable and available from multiple sources.  

You might calculate the multihash from the source code of the parser
or handler for the protocol.  This is useful for protocols that are
not otherwise formally specified, and it takes advantage of the
precision of a programming language rather than a human-language
specification. It does tie the multihash to an exact version of the
implementation, which has both advantages and disadvantages.

You should publish the multihash and document reference for any
protocol you design or use. A simple way to do this is to add the
hex-encoded multihash, along with a citation for the protocol document
it refers to, in a source code comment in your handler or dispatcher.

