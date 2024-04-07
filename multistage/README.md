# MultiStage (Multi Pass) Translator, Compiler, and Interpreter Framework

Translating an input file or stream into an output is a common task in
computer science.  In multistage_test.go we show a simple technique
for doing this using a series of stages.  Each stage is a function
that takes an input channel and returns an output channel.  The output
of one stage is the input to the next stage.  Because the I/O is via
channels, the stages can run concurrently in a pipeline if needed.

Backtracking, where a stage needs to re-read input, is commonly used
in parsing.  The Backtracker type is a simple generic backtracking
buffer that can be used to implement backtracking in a stage.  A stage
imports the backtracker and uses it checkpoint and rollback messages
it receives from the input channel.

While the idiomatic way to write libraries in Go is to not use
channels as part of the API, this is a case where channels may make
sense. An alternative might be to use a Stage interface with methods
like Start, Next, and Close.  This would be more idiomatic, but would
require more boilerplate code to use.

The architecture I'm converging toward appears to be:
- A stage is a function that takes an input channel and returns an output channel.
- Internally, a stage can use a backtracker to implement backtracking.
  This framework can provide a backtracker type to make this easier.
- Internally, a stage uses recursive descent to parse its input.  This
  framework can provide a Try function to make this easier.  The Try
  function would use the backtracker to implement backtracking.

An alternative to the backtracker might be to use a parser combinator
library. The reason a parser combinator library doesn't need
backtracking is that it uses a different parsing technique.  Instead
of using recursive descent, it uses a technique called "packrat
parsing" which is a form of memoization. Packrat parsing works by
using a map to cache the results of parsing subexpressions.  We could
instead use channels and goroutines to implement packrat parsing,
where each goroutine is essentially doing speculative parsing of a
subexpression.  This would be a more idiomatic way to use channels in
Go, but would be more complex than the backtracker approach.

Other or hybrid approaches might be possible.  For example, we could
use channels to connect goroutine stages, but instead of a single
input channel and a single output channel, we could use a data channel
and a control channel.  The control channel could be used to send
messages like "checkpoint" and "rollback" to the previous stage.  This
would be more idiomatic that a backtracker, but would be more complex
than a simple channel-based pipeline.  We could also use a dataflow
programming model, where stages are connected by channels, and the
channels are connected by a dataflow graph.  This would be more
complex than a simple pipeline, but would be more flexible.

XXX TRUE? An interesting aspect of packrat parsing is that it is a
form of dynamic programming.  This means that it can be used to solve
optimization problems.  For example, we could use packrat parsing to
implement a dynamic programming algorithm to solve the longest common
subsequence problem.  This would be a good example of a stage that
does not have a simple input/output relationship with the previous
stage.  Instead, it would have a more complex relationship, where it
needs to read the entire input before it can produce any output.

Another interesting aspect of packrat parsing is the memoization
cache.  This cache allows us to avoid re-parsing the same subexpression
multiple times.  I.e., it treats the process as referentially
transparent.  This is similar to the core of PromiseGrid, where we
treat functions as referentially transparent.  This suggests that we
could use packrat parsing to implement PromiseGrid or vice versa.

I suspect that the PromiseGrid dispatcher/cache is the right kernel.
In that design, the cache is a cache of promises, where a cache entry
is a promise that the value or function pointer is correct.  The cache
is a map from a key sequence to a promise.  The dispatcher is a
function that takes a key sequence and returns a promise.  



