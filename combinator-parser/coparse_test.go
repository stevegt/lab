package coparse

import (
	"context"
	"strings"

	"testing"

	. "github.com/stevegt/goadapt"
)

type PassthroughT struct {
	NodeBase
}

func PassThrough(ctx context.Context, src []byte) *PassthroughT {
	n := &PassthroughT{}
	n.FromBytes(src)
	return n
}

func (n *PassthroughT) Parse(ctx context.Context) bool {
	n.Start = 0
	n.End = len(n.src)
	return true
}

type EditT struct {
	NodeBase
	fn func(string) string
}

func Edit(ctx context.Context, src []byte, fn func(string) string) *EditT {
	n := &EditT{}
	n.FromBytes(src)
	n.fn = fn
	return n
}

func (n *EditT) Parse(ctx context.Context) bool {
	n.Start = 0
	n.End = len(n.src)
	n.
	XXX render, value, src?  what do these mean?
	return true
}

func TestHelloWorld(t *testing.T) {
	src := "Hello, World!"
	top := PassThrough(context.Background(), []byte(src))
	Tassert(t, top.Parse(context.Background()))
	Tassert(t, top.String() == src)

	lower := func(s string) string {
		return strings.ToLower(s)
	}
	edit := Edit(context.Background(), lower)



	Try(context.Background(), top, edit)

}

/*
func TestHelloWorld(t *testing.T) {
	ctx := context.Background()
	Try(ctx, func(ctx context.Context) error {



	// Collect and print the result
	for result := range output2 {
		fmt.Println(result)
		Tassert(t, result == "hello, big world!")
	}
}


// Lower converts string to lowercase
func Lower(input <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for s := range input {
			out <- strings.ToLower(s)
		}
		close(out)
	}()
	return out
}

// Big inserts "big" before "world"
func Big(input <-chan string) <-chan string {
	out := make(chan string)
	go func() {
		for s := range input {
			if strings.Contains(s, "world") {
				out <- strings.Replace(s, "world", "big world", -1)
			} else {
				out <- s
			}
		}
		close(out)
	}()
	return out
}

// String2Message converts string to message
func String2Message(input <-chan string) <-chan Message {
	out := make(chan Message)
	go func() {
		for s := range input {
			msg := NewDefaultMessage()
			msg.SetText(s)
			out <- msg
		}
		close(out)
	}()
	return out
}

// DebugMessages prints the message stream
func DebugMessages(input <-chan Message) <-chan Message {
	out := make(chan Message)
	go func() {
		for msg := range input {
			Pf("%#v\n", msg)
			out <- msg
		}
		close(out)
	}()
	return out
}

func TestRegex(t *testing.T) {
	input := Start("Hello, World!")

	// Create a regular expression pass to match "Hello"
	hello := Regex(`^Hello`, NewDefaultMessage)

	msgs := String2Message(input)
	msgs = DebugMessages(msgs)
	output := Stage(msgs, hello)

	for result := range output {
		fmt.Println(result.Text())
		Tassert(t, result.Text() == "Hello")
	}
}
*/

/*
// user-defined token types
const (
	Word       = TokenType("Word")
	Whitespace = TokenType("Whitespace")
)

func TestToken(t *testing.T) {
	// Start the pipeline with the initial input
	input := Start("Hello, World!")

	// Create a tokenizing pass
	tokens := Token{
		Rules: []Rule{
			{Pattern: `\S+`, Type: Word},
			{Pattern: `\s+`, Type: Whitespace},
		},
	}

	// Start the goroutine for the tokenizing pass
	output := tokens.Run(input)

	// Collect and test the result
	for token := range output {
		fmt.Println(token)
	}
}


const (
	Word       = TokenType("Word")
	Whitespace = TokenType("Whitespace")
)

// Token tokenizes the input string
type Token struct{
	Type string
	Value string
	Line int
	Column int
}

func (p *Token) Run(input <-chan string) <-chan Token {
	out := make(chan Token)
	go func() {
		line := 1
		// split the input into lines
		lines := strings.Split(<-input, "\n")
		for l, line := range lines {
			// split the line into tokens
			t := &Token{Line: l}
			for c, b := range []byte(line) {
				t.Column = c
				tail := line[c:]

				switch {
				case strings.HasPrefix(tail, "hello"):
					t.Type = "greeting"
				case strings.HasPrefix(tail, "world"):
					t.Type = "location"
				default:

					t.Type = "unknown"
		}
		close(out)
	}()
	return out
}

func TestToken(t *testing.T) {
	// Start the pipeline with the initial input
	input := Start("Hello, World!")

	// Create a tokenizing pass
	tokens := Token{
		Rules: []Rule{
			{Pattern: `\S+`, Type: Word},
			{Pattern: `\s+`, Type: Whitespace},
		},
	}

	// Start the goroutine for the tokenizing pass
	output := tokens.Run(input)

	// Collect and test the result
	for token := range output {
		fmt.Println(token)
	}
}

*/
