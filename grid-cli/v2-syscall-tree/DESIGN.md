# PromiseGrid Design

## Overview

This document outlines the design considerations and architecture for implementing PromiseGrid.

## Core Concepts

1. **Decentralized Architecture**: PromiseGrid operates as a
   decentralized computing, communications, and governance system. It
   is designed to be owned and operated by its users rather than any
   single entity.

2. **Capability-as-Promise Model**: Capabilities are treated as
   promises, similar to the concepts from Promise Theory. A capability
   token represents a promise that can either be fulfilled or revoked.

3. **Content-addressable Code**: Both code and data are addressed by
   their content, not by location or name. This allows the grid to
   store and execute code and access data from any node in the
   network.

4. **Promises All the Way Down**: Every interaction in the system is
   based on promises. A response to a promise is another promise.

## Message Structure

- The `Message` structure includes the promise as the first element in
  the `Parms` field.  Recipients route or discard messages based on the
  promise.
  
```go
type Message struct {
    Parms    []interface{}          `json:"parms"`    // Parameters, with promise as the first element
    Payload  map[string]interface{} `json:"payload"`  // Meta information or additional data
}
```

## Syscall Tree

- **Hierarchical Syscall Tree**: The kernel uses a hierarchical
  syscall tree to store acceptance history. This tree functions as an
  "ant routing" mechanism, caching successful paths to optimize future
  routing.

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

- **Single `Accept()` Function**: The `Module` interface includes a single `Accept()` function that returns a promise message. The kernel routes the message to the module whose syscall tree key matches the most leading parameter components.

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
    // Implement logic to accept or reject based on parms
    return Message{Parms: append([]interface{}{true}, parms...), Payload: map[string]interface{}{"info": "cache module"}}, nil
}

func (m *LocalCacheModule) HandleMessage(ctx context.Context, parms ...interface{}) ([]byte, error) {
    // Implement logic to handle messages
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

- **Promises and Accountability**: The acceptance of a message itself
  is a promise. Modules track which requests they accept and must
  fulfill these promises by successfully handling the requests.

## Integration with WebSocket

- **WebSocket Handling**: The kernel interacts with modules through
  WebSocket connections, routing and filtering messages based on the
  hierarchical syscall tree.

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

## Summary

- **Unified Message Structure**: Integrating promises directly into the message parameters.
- **Hierarchical Syscall Tree**: Ensuring efficient routing and caching based on successful paths.
- **Promise-Based Acceptance**: Treating acceptance as a promise, ensuring modules fulfill their commitments.
- **Integration with WebSocket**: Providing a consistent interface for handling and routing messages in a decentralized manner.

This design aligns with the principles of decentralized governance, modular interaction, and trust-based operations, ensuring a robust and efficient system for PromiseGrid.

