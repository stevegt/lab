
from CSWG Office Hours 05 Sep 2023

* recap of current toygrid status
    * next dev step is nailing down the message format and transport for interbrowser comms (e.g. db app updates)
    * simple thing would be to use MQTT [https://en.wikipedia.org/wiki/MQTT](https://en.wikipedia.org/wiki/MQTT) 
        * but MQTT uses a central server
    * Steve would prefer to not build a reliance on MQTT, even as early as toygrid, because once started, aside from needing to avoid embedding MQTT client code too deeply, the assumptions built into a pub/sub messaging system like MQTT are hard to eradicate later
    * Steve started work on defining the messaging format and transport, and that has developed into the “white paper” work above
* Q: does Steve press ahead with the white paper work, or go back and punt with MQTT to get a toygrid demo working sooner?
    * Donaldo:  
        * press ahead with white paper
        * if/when interacting with IoP folks, e.g. in a demo, what is important?
* Q:  if/when demoing to IoP folks, what is more important?
    * showing them a toygrid that is also centralized (MQTT/Yjs) just like mypads.js, and just uses tiptap instead of etherpad?  (not much gain -- not very interesting unless we beef up the apps, but even if WASM they aren’t decentralized)
    * showing them a thing that is decentralized, works like mypads/etherpad, but doesn’t depend on a single server or legal entity to maintain?
    * the latter is more sustainable and is more likely to align with IoP philosophy
* Donaldo:  my wish is that toygrid is online somewhere full-time so it can be discussed ad-hoc whenever -- something to point people at in conversation, to work on the UI further, get feedback on the UI/UX.
    * if toygrid is centralized:  mypads or etherpad already is an editor, but doesn’t have apps, so we’re basically just  demonstrating WASM
    * if toygrid is decentralized: we’re demonstrating both WASM and decentralization
* Steve giving Donaldo a quick trip down the rabbit hole of the draft white paper:
    * now (toygrid):
        * static server serves code and HTML
        * Yjs server stores state
        * websocket server relays messages and is stateless
    * future (toygrid or ?):
        * static servers in a DNS pool serve code and HTML
            * static content must be kept small and easily verifiable
            * static content may be sourced from grid data in infrequent releases
        * websocket servers in a DNS pool relay messages and are stateless
    * state: content of all docs, files, media
    * transition: change of state caused by an actor or function
    * side effects: 
        * TODO:  Steve continue here with Donaldo
        * influence between a transition and the external universe
            * in either direction
        * possibly non-reproducible, non-idempotent, harmful if repeated
            * example: transition caused a motor to turn
            * example: getting data from user (repeatedly)
    * decentralization depends on any host being able to “replay” transitions from any state to next state, to create and/or verify current state
        * but replaying a transition that had side effects is not safe
        * so for those transitions that are known to have side effects, we instead have the machine that executed the transition provide a signed message asserting that the transition took place and here’s the hash of the new state
    * but how do we know when a transition has a side effect?
        * runtime executes all transition functions in a sandbox (e.g. WASM VM)
            * so they can’t reach the outside world
            * we provide a signed message with the old and new state hashes, transition hash, hash of arg set
        * runtime keeps state (content) in local CAS, keyed by hash
            * runtime assumes all CAS reads (and only CAS reads) have side effects
            * example: the only way a motor can know it needs to turn is if its controller reads a “turn” command in the new state; the code that generated the “turn” command ran inside the sandbox and couldn’t reach the outside world so had no direct side effect
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

