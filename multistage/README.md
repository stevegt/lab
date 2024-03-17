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
