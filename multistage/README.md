# MultiStage (Multi Pass) Translator, Compiler, and Interpreter Technique


Translating an input file or stream into an output is a common task in
computer science.  In main.go we show a simple technique for doing
this using a series of stages.  Each stage is a function that takes an
input channel and returns an output channel.  The output of one stage
is the input to the next stage.  Because the I/O is via channels, the
stages can run concurrently in a pipeline if needed.


