# MultiPass Translator, Compiler, and Interpreter Framework

Translating an input into an output is a common task in computer
science. This framework provides a way to define a series of passes
that can be used to translate an input into an output. Each pass is a
function that takes an input channel and returns an output channel.
The output of one pass is the input to the next pass.  Because the I/O
is via channels, the passes can run concurrently in a pipeline if
needed. 

## Example

Here is an example of how to use this framework to define a series of
passes and then run them in order:

```go
package main

import (
    "fmt"
    "strings"
    "unicode"
)

// Passes must implement the Pass interface
type Pass interface {
    Run(input <-chan any) (output <-chan any)
}

// Pass1 converts the input to upper case
type Pass1 struct{}

func (p *Pass1) Run(input <-chan string) (output <-chan string) {
    out := make(chan string)
    go func() {
        for s := range input {
            out <- strings.ToUpper(s)
        }
        close(out)
    }()
    return out
}

// Pass2 replaces "!" with "?"
type Pass2 struct{}

func (p *Pass2) Run(input <-chan string) (output <-chan string) {
    out := make(chan string)
    go func() {
        for s := range input {
            out <- strings.Replace(s, "!", "?", -1)
        }
        close(out)
    }()
    return out
}


func main() {
    input := "Hello, World!"
    passes := []func(string) string{
        strings.ToUpper,
        func(s string) string {
            return strings.Replace(s, "!", "?", -1)
        },
        func(s string) string {
            return strings.Map(func(r rune) rune {
                if unicode.IsLetter(r) {
                    return r + 1
                }
                return r
            }, s)
        },
    }
    output := multipass.Pipeline(input, passes)
    fmt.Println(output)
}
```


