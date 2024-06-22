# Simplified Handling of Parameters and Acceptance History with Promises

## Overview
This document outlines the implementation strategy for handling messages based on promises, module routing, and optimization using a hierarchical syscall tree, focusing on the concept of promises and the unification of *Accept* and *HandleMessage* functionalities.

## Implementation Strategy

### Unified Message Structure

- **Message Structure**: Incorporate the promise directly as the first element in the `Parms` field of the `Message`.
- **Structure Definition**:
  ```go
  type Message struct {
      Parms    []interface{}          `json:"parms"`
      Payload  map[string]interface{} `json:"payload"`
  }
  ```

### Hierarchical Syscall Tree (Ant Routing)

- **Cached successful paths**: Enhance routing efficiency for future calls with similar parameters.
- **Tree Structure**: Each node represents a level of parameter matching, with leaf nodes potentially corresponding to specific modules or cached results.
    ```go
    type SyscallNode struct {
        modules  []Module
        children map[string]*SyscallNode
    }

    type Kernel struct {
        root *SyscallNode
        knownMessages map[string]Message
    }
    ```

### Simplifying Accept and HandleMessage into a Promise-Based Unified Approach

- **Promise-Based Decision**: Consolidate *Accept* and *HandleMessage* into one step, where the promise to process a message is fulfilled immediately, and a promise message is generated and returned.
- **PromiseMessage as Message**: Treat the *PromiseMessage* as merely the first element of `Parms`, introducing a seamless contract between caller and module.
- **Dynamic Sys.

#### Promise-Based Acceptance Mechanism

- **Acceptance as a Promise**: The first element in `Parms` indicates whether the module promises to handle the request.
    - **Accept Function**: 
      ```go
      func (m *Module) Accept(ctx context.Context, parms ...interface{}) (Promise, error)
      ```
    - **Handling Function**: 
      ```go
      func (m *Module) HandleMessage(ctx context.Context, parms ...interface{}) ([]byte, error)
      ```

### Consulting Modules with Promise-Based Acceptance

- **Optimized Module Consultation**: When a message is received, consult the syscall tree to find the longest matching parameter slice and route the message to the respective module.
- **Caching Successful Calls**: Use the syscall tree to cache paths of successful module consultations, akin to ant routing, where a trail is left for future use.

### Promise Message and Meta-Information Handling

- **Unified Message Handling**: Integrate additional meta-information (such as trust levels, module capabilities, etc.) into the Promise message returned by the Accept or Handling process.
- **Meta as Message.Payload**: Include any metadata or additional guarantees about the promise in the `Message.Payload`.

```go
func handleWebSocket(ctx context.Context, k *Kernel, w http.ResponseWriter, r *http.Request) { ... }
```

### Discussion: Integration with "Promises All the Way Down"

- **Trust and Accountability with Promise-Based Design**: In this design, every interaction within the system is based on trust represented by promises. The acceptance and fulfillment of these promises form the foundation of module interactions, enhancing system resilience and adaptability.

### Conclusion

This approach of simplifying message handling and acceptance criteria into a unified, promise-based mechanism not only streamlines interactions within the system but also aligns with foundational computing theory principles, ensuring a robust framework for decentralized cooperation and governance.

