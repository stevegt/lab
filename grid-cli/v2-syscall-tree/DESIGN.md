# PromiseGrid Design

## Overview

This document outlines the design considerations and architecture for implementing PromiseGrid.

## Core Concepts

1. **Decentralized Architecture**: PromiseGrid operates as a decentralized computing, communications, and governance system. It is designed to be owned and operated by its users rather than any single entity.

2. **Capability-as-Promise Model**: Capabilities are treated as promises, similar to the concepts from Promise Theory. A capability token represents a promise that can either be fulfilled or revoked.

3. **Content-addressable Code**: Both code and data are addressed by their content, not by location or name. This allows the grid to store and execute code and access data from any node in the network.

4. **Promises All the Way Down**: Every interaction in the system is based on promises. A response to a promise is another promise.

## Message Structure

- The `Message` structure includes the promise as the first element in the `Parms` field. Recipients route or discard messages based on the leading promise.

```go
type Message struct {
    Parms    []interface{}          `json:"parms"`    // Parameters, with promise as the first element
    Payload  map[string]interface{} `json:"payload"`  // Meta information or additional data
}
```

## Syscall Tree

- **Hierarchical Syscall Tree**: The kernel uses a hierarchical syscall tree to store acceptance history. This tree functions as an "ant routing" mechanism, caching successful paths to optimize future routing.

```go
type SyscallNode struct {
    modules  []Module
    children map[string]*SyscallNode
}

type Kernel struct {
    root          *SyscallNode
    knownMessages map[string]Message
}
```

- **Dynamic Acceptance History**: The syscall tree captures positive and negative acceptance history. It starts empty and is populated during operation as the kernel consults built-in and other modules to handle received messages.

## Module Interface

- **`Accept()` Function**: The `Module` interface includes an `Accept()` function that returns a promise message. The returned promise is a promise that the module can handle the message. The kernel routes the message to the module whose syscall tree key matches the most leading parameter components.

```go
type Module interface {
    Accept(ctx context.Context, parms ...interface{}) (Message, error)
    HandleMessage(ctx context.Context, parms ...interface{}) ([]byte, error)
}

type LocalCacheModule struct {
    cacheDir string
}

func NewLocalCacheModule(cacheDir string) *LocalCacheModule {
    return &LocalCacheModule{cacheDir: cacheDir}
}

func (m *LocalCacheModule) Accept(ctx context.Context, parms ...interface{}) (Message, error) {
    // Implement logic to accept or reject based on parms.
    return Message{Parms: append([]interface{}{true}, parms...), Payload: map[string]interface{}{"info": "cache module"}}, nil
}

func (m *LocalCacheModule) HandleMessage(ctx context.Context, parms ...interface{}) ([]byte, error) {
    // Implement logic to handle messages.
}
```

## Routing and Filtering

- **Optimized Routing**: In the case of a cache miss, the kernel consults modules based on the hierarchical syscall tree. It routes the message to the module with the longest matching parameter slice.

```go
func (k *Kernel) consultModules(ctx context.Context, parms ...interface{}) ([]byte, error) {
    bestMatch := k.findBestMatch(parms...)
    var promisingModules []Module

    for _, module := range bestMatch.modules {
        promise, err := module.Accept(ctx, parms...)
        if err != nil {
            continue // Log and handle error
        }
        if promise.Parms[0].(bool) {
            promisingModules = append(promisingModules, module)
            k.addSyscall(parms...) // Add to syscall tree
        }
    }

    for _, module := range promisingModules {
        data, err := module.HandleMessage(ctx, parms...)
        if err == nil {
            return data, nil
        }
    }

    return nil, fmt.Errorf("no module could handle the request")
}
```

## Acceptance as a Form of Promise

- **Promises and Accountability**: The acceptance of a message itself is a promise. Modules track which requests they accept and must fulfill these promises by successfully handling the requests.
- The use of "accept" in this context aligns with the definitions in computing theory [^church][^turing][^chomsky]: An automaton accepts an input if it reaches an accepting state. Similarly, PromiseGrid modules accept a message if they can handle it, making a promise to process it, akin to how a Turing machine or a language automaton accepts strings belonging to a language.

## Integration with WebSocket

- XXX nodes interact with peers over the network via WebSocket connections
- XXX decide whether websocket is the only transport or if we can have others
- XXX decide how a sandboxed module can interact with the network
- XXX decide whether the kernel should handle the WebSocket connection or if it should be handled by non-sandboxed modules


```go
func handleWebSocket(ctx context.Context, k *Kernel, w http.ResponseWriter, r *http.Request) {
    upgrader := websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
    }
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        return
    }
    defer conn.Close()

    for {
        _, message, err := conn.ReadMessage()
        if err != nil {
            break
        }

        parms, err := deserializeMessage(message)
        if err != nil {
            continue
        }

        data, err := k.consultModules(ctx, parms...)
        if err != nil {
            continue
        }

        if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
            break
        }
    }
}
```

## Open Questions

- What are the specific conditions under which the kernel should
  invalidate or update cached syscall paths in the hierarchical
  syscall tree?
- How can we dynamically adjust the acceptance criteria of modules to
  adapt to changing workloads and conditions without manual
  intervention?
- What mechanisms can be implemented to handle broken promises more
  effectively, ensuring minimal disruption to the system?
- Should there be a standardized format for the payload in promise
  messages to facilitate easier parsing and validation across
  different modules?
- Regarding the design choice of using a separate Accept() and
  Handle() method -- does this not break promise theory's principle of
  not making promises on behalf of others?  I.E. if there is a
  separate Accept() and Handle() method, this means that the Accept()
  code path is making a promise on behalf of the Handle() code path.
  What are the implications of this?  Should this design be changed?


## References

[^church]: XXX some reference to Church's work on lambda calculus

[^turing]: Turing, Alan. "On computable numbers, with an application to the Entscheidungsproblem." Proceedings of the London Mathematical Society 2.1 (1937): 230-265.  XXX add URL

[^chomsky]: Chomsky, Noam. "Three models for the description of language." IRE transactions on information theory 2.3 (1956): 113-124. XXX add URL
