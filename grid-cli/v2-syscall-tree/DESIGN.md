# PromiseGrid Design

## Overview

This document outlines the design considerations and architecture for implementing PromiseGrid.

## Core Concepts

1. **Decentralized Architecture**: PromiseGrid operates as a decentralized computing, communications, and governance system. It is designed to be owned and operated by its users rather than any single entity.

2. **Capability-as-Promise Model**: Capabilities are treated as promises, similar to the concepts from Promise Theory. A capability token represents a promise that can either be fulfilled or revoked.

3. **Content-addressable Code**: Both code and data are addressed by their content, not by location or name. This allows the grid to store and execute code and access data from any node in the network.

4. **Promises All the Way Down**: Every interaction in the system is based on promises. A response to a promise is another promise.

5. **Non-Sandboxed Modules**: Non-sandboxed modules in PromiseGrid are analogous to device drivers in a microkernel OS. Just as device drivers handle specific hardware functionality in a microkernel, non-sandboxed modules handle specific external operations in the grid (e.g., network communications, file access). The kernel delegates these operations to non-sandboxed modules while maintaining control over the overall execution.

## Message Structure

- The `Message` structure includes the promise as the first element in the `Parms` field. Recipients route or discard messages based on the leading promise.

## Syscall Tree

- **Hierarchical Syscall Tree**: The kernel uses a hierarchical syscall tree to store acceptance history. This tree functions as an "ant routing" mechanism, caching successful paths to optimize future routing.


- **Dynamic Acceptance History**: The syscall tree captures positive and negative acceptance history. It starts empty and is populated during operation as the kernel consults built-in and other modules to handle received messages.

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
        acceptable, data, err := module.HandleMessage(ctx, true, parms...)
        if err != nil || !acceptable {
            continue
        }
        // Now handle the message
        _, data, err = module.HandleMessage(ctx, false, parms...)
        if err == nil {
            return data, nil
        }
    }

    return nil, fmt.Errorf("no module could handle the request")
}
```

## Acceptance as a Form of Promise

- **Promises and Accountability**: The acceptance of a message itself is a promise. Modules track which requests they accept and must fulfill these promises by successfully handling the requests.
- The use of "accept" in this context aligns with the definitions in computing theory: An automaton accepts an input if it reaches an accepting state. Similarly, PromiseGrid modules accept a message if they can handle it, making a promise to process it, akin to how a Turing machine or a language automaton accepts strings belonging to a language. 

## Integration with WebSocket

- Nodes interact with peers over the network via WebSocket connections.
- WebSocket is the message transport mechanism we're using for now, although other mechanisms may be adopted in the future.
- A sandboxed module can interact with the network by sending and receiving messages through the kernel.
- The kernel communicates with the outside world (both network and local I/O) via non-sandboxed modules.  

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

## Glossary of Terms and Concepts

- **Decentralized Architecture**: A system design where control is distributed among various actors or nodes, rather than centralized in a single entity.
- **Capability-as-Promise Model**: A model where capabilities represent promises that can either be fulfilled or revoked.
- **Content-addressable Code**: Code and data addressed based on their content, allowing decentralized storage and execution.
- **Promises All the Way Down**: A paradigm where every interaction in the system results in a promise, with responses being new promises.
- **Non-Sandboxed Modules**: Modules granted more access and responsibility, analogous to device drivers, for handling specific external operations.

## Open Questions

- What are the specific conditions under which the kernel should invalidate or update cached syscall paths in the hierarchical syscall tree?
- How can we dynamically adjust the acceptance criteria of modules to adapt to changing workloads and conditions without manual intervention?
- What mechanisms can be implemented to handle broken promises more effectively, ensuring minimal disruption to the system?
- Regarding the design choice of using a separate `Accept()` and `HandleMessage()` method -- does this not break promise theory's principle of not making promises on behalf of others? If there is a separate `Accept()` and `HandleMessage()` method, this means that the `Accept()` code path is making a promise on behalf of the `HandleMessage()` code path. What are the implications of this? Should this design be changed?
- How does the kernel determine the best route when multiple modules provide promises to handle a message?
- How will the system manage latency in the context of decentralized and distributed nodes?
- What measures can be taken to ensure data consistency across the decentralized cache system over unreliable networks?

## Suggestions for Improving this Document

- Develop comprehensive error handling and logging for broken promises to ensure accountability and system robustness.
- Add more examples and case studies to illustrate core principles and use cases.
- Expand on the description of the hierarchical syscall tree and its role in routing and filtering messages.
- Update the sections with any additional questions or areas of exploration that arise during implementation and testing.
- Refine the glossary to include more specialized terms and concepts as they arise in the documentation.
- Incorporate visual aids (e.g., diagrams, flowcharts) to enhance understanding of system architecture and data flow.
