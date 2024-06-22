# TODO List for Implementing PromiseGrid Kernel 

## Simplify Message Structure
- Implement a unified `Message` structure where the promise might be the first element in the `Parms` field.

## Module Interface
- Simplify the Module interface to include a single `Accept()` function that handles all acceptance criteria based on the parameters (`parms...`).

## Hierarchical Syscall Tree
- Implement a hierarchical syscall tree as an "ant routing" mechanism to cache successful paths and optimize future routing.
- Incorporate promise-based acceptance in the syscall tree to enhance dynamic routing and module selection.

## Unified PromiseMessage
- Treat `PromiseMessage` as a simple `Message` where the first element of `Parms` can be the promise.
- Implement logic to handle `PromiseMessage` as part of the acceptance and handling process.

## Implement Ant Routing Mechanism
- Incorporate ant routing mechanisms to cache the path to modules that have successfully handled previous calls.
- Use the syscall tree to route messages to the module whose syscallTable key matches most leading `parms` components.

## Dynamic Acceptance History and Syscall Table
- Ensure the `acceptanceHist` and `syscallTable` are unified, serving as a dynamic syscall table that records positive and negative acceptance history.
- Populate the dynamic syscall table during kernel operation as it consults first builtins and then other modules to handle received messages.
- Consider storing the dynamic syscall table as a form of acceptance history in an embedded data structure for initial module recognition.

## Promise-Based Design
- Deepen the "promises all the way down" design by ensuring that even the `Accept()` response is integrated as a form of promise.
- Explore the possibility of making the `Accept()` response itself a promise message, potentially including meta-information within the `Message.Payload`.

## Error Handling and Promise Fulfillment
- Develop mechanisms to handle situations where a module's promise to handle a message isn't fulfilled due to failure in `HandleMessage()`.
- Consider implementing fallback or retry mechanisms for failed promise fulfillment to enhance system resilience and reliability.
