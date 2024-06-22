# TODO List for Implementing PromiseGrid Kernel 

## Hierarchical Syscall Tree
- Implement a hierarchical syscall tree as an "ant routing" mechanism to cache successful paths and optimize future routing.
- Incorporate promise-based acceptance in the syscall tree to enhance dynamic routing and module selection.

## Implement Ant Routing Mechanism
- Incorporate ant routing mechanisms to cache the path to modules that have successfully handled previous calls.
- Use the syscall tree to route messages to the module whose syscallTable key matches the most leading `parms` components.

## Unified Message Structure
- Define a unified `Message` structure where the promise is the first element in `Parms`, potentially including meta-information in `Payload`.
- Ensure that the message format allows for the leading hash to be used as a promise hash, followed by the module hash and arguments.

## Cache Design and Lookup
- Ensure that the cache uses filesystem separators (`/`) between each key component and URL-encodes the arguments.
- Treat modules as caches and consult them on cache misses, potentially using multiple caches including the built-in kernel cache and module-provided caches.
- Implement a complex filter algorithm to route messages based on promise acceptance, module hash match, and argument acceptance.

## Dynamic Acceptance History and Syscall Table
- Merge `acceptanceHist` and `syscallTable` into a single dynamic syscall table.
- Populate the dynamic syscall table during kernel operation, starting with built-in modules and then consulting other modules as needed.
- Store positive and negative acceptance history for all modules to optimize future lookups.

## Promise-Based Design
- Deepen the "promises all the way down" design by treating acceptance as a promise message.
- Explore the possibility of using the `Accept()` response itself as a promise message, with meta-information included in `Payload`.
- Ensure that the kernel includes known accept/reject messages in its embedded data for validation and recognition.

## Error Handling and Promise Fulfillment
- Log and handle broken promises when `HandleMessage()` fails after `Accept()` returned true.
- Develop mechanisms to reroute requests or de-prioritize unreliable modules to maintain system resilience and trust.
- Track and validate acceptance promises to ensure accountability and trustworthiness within the system.
