package main

import (
	"regexp"
	// . "github.com/stevegt/goadapt"
)

/*
MultiPass/Multi-Stage Translator, Compiler, and Interpreter Framework

Translating an input into an output is a common task in computer
science. This framework provides a way to define a series of stages
that can be used to translate an input into an output. Each stage is a
function that takes an input channel and returns an output channel.
The output of one stage is the input to the next stage.  Because the I/O
is via channels, the stages can run concurrently in a pipeline if
needed.

Each stage can be implemented as a function or a method. The Go type
system ensures that the input and output types are compatible.

Each stage might split or aggregate the input; for instance, a
tokenizer stage might split the input into tokens, and a parser stage
might aggregate the tokens into a parse tree.

This framework includes a default implementation of a combinator stage.

*/

// Start can be used to start a pipeline with one or more input
// values, with each value being the contents of a file, a string,
// struct, or any other type.  The optional combinator stage uses
// Message as its input and output interface.
func Start[T any](values ...T) <-chan T {
	out := make(chan T)
	go func() {
		for _, value := range values {
			out <- value
		}
		close(out)
	}()
	return out
}

// Message is an optional convenience interface for representing an item
// being communicated between stages.  It can be used to represent a
// token, a parse tree, an abstract syntax tree, or any other
// intermediate representation. Message is used by the included
// combinator stage, but could be used by any stage that needs to
// communicate structured data with position information.
type Message interface {
	Parent() Message
	SetParent(Message)
	Children() []Message
	AddChild(Message)
	Position() (filename string, line int, column int)
	SetPosition(filename string, line int, column int)
	Text() string
	SetText(string)
}

type NewMessage func() Message

// DefaultMessage is a default implementation of the Item interface
type DefaultMessage struct {
	parent   Message
	children []Message
	filename string
	line     int
	column   int
	text     string
}

func NewDefaultMessage() Message {
	return &DefaultMessage{}
}

// Parent returns the parent message
func (m *DefaultMessage) Parent() Message {
	return m.parent
}

// SetParent sets the parent message
func (m *DefaultMessage) SetParent(parent Message) {
	m.parent = parent
}

// Children returns the children messages
func (m *DefaultMessage) Children() []Message {
	return m.children
}

// AddChild adds a child message
func (m *DefaultMessage) AddChild(child Message) {
	m.children = append(m.children, child)
}

// Position returns the filename, line, and column of the message
func (m *DefaultMessage) Position() (filename string, line int, column int) {
	return m.filename, m.line, m.column
}

// SetPosition sets the filename, line, and column of the message
func (m *DefaultMessage) SetPosition(filename string, line int, column int) {
	m.filename = filename
	m.line = line
	m.column = column
}

// Text returns the text of the message
func (m *DefaultMessage) Text() string {
	return m.text
}

// SetText sets the text of the message
func (m *DefaultMessage) SetText(text string) {
	m.text = text
}

// EOF represents the end of the input
type EOF struct{ DefaultMessage }

// Unknown represents an unknown item
type Unknown struct{ DefaultMessage }

func Stage(input <-chan Message, fn Parser) <-chan Message {
	output := make(chan Message)
	go func() {
		var buffer []Message
		var best *Result

		// flush the consumed messages and send the best result to the output channel
		flush := func() {
			if best != nil {
				buffer = buffer[best.Consumed:]
				for _, msg := range best.Messages {
					output <- msg
				}
				best = nil
			}
		}

		for msg := range input {
			// accumulate messages in the buffer
			buffer = append(buffer, msg)

			// Attempt to parse the current buffer.
			res := fn(buffer)
			if res.Matched {
				// If there's a match, save the result and wait for more tokens to see if the match can be extended.
				best = &res
			} else if best != nil {
				// If the current token breaks the match and there's a stashed result, flush the stashed result.
				flush()
			}
		}

		// After all tokens are processed, flush any stashed result.
		flush()

		// If there are remaining tokens that didn't contribute to a
		// final match, bundle them into an Unknown message and send
		// that to the output channel.
		if len(buffer) > 0 {
			unk := &Unknown{DefaultMessage{children: buffer}}
			best = &Result{
				Messages: []Message{unk},
				Consumed: len(buffer),
			}
			flush()
		}
		close(output)
	}()
	return output
}

// Result is a type that represents the result of a parsing operation.
type Result struct {
	// boolean indicating whether the parsing operation matched the input.
	Matched bool
	// output messages produced by the parsing operation.
	Messages []Message
	// number of input messages consumed by the parsing operation.
	Consumed int
}

// Parser is a function type that takes input messages and returns
// a Result.
type Parser func([]Message) Result

// Combinator is a function that takes parsers and returns a new parser.
type Combinator func(...Parser) Parser

// Sequence is a combinator that takes a sequence of parsers and
// returns a new parser that matches the sequence.
func Sequence(parsers ...Parser) Parser {
	return func(input []Message) Result {
		var output Result
		for _, parser := range parsers {
			res := parser(input)
			if !res.Matched {
				return output
			}
			output.Messages = append(output.Messages, res.Messages...)
			output.Consumed += res.Consumed
		}
		output.Matched = true
		return output
	}
}

// Or is a combinator that takes a sequence of parsers and returns a
// new parser that matches the first successful parser.
func Or(parsers ...Parser) Parser {
	return func(input []Message) Result {
		for _, parser := range parsers {
			res := parser(input)
			if res.Matched {
				return res
			}
		}
		return Result{}
	}
}

// Regex returns a parser.  The parser matches a regular expression in
// the first message it is given and returns a new message containing
// the matched text.  If a submatch is specifed in the regex,
// the submatch content is used as the output value.
func Regex(pattern string, constructor NewMessage) Parser {
	re := regexp.MustCompile(pattern)
	return func(input []Message) Result {
		if len(input) == 0 {
			return Result{}
		}
		msg := input[0]
		txt := msg.Text()
		m := re.FindStringSubmatchIndex(txt)
		if m == nil {
			return Result{}
		}
		// matched
		var value string
		if len(m) > 2 {
			// use first submatch as the output value
			value = txt[m[2]:m[3]]
		} else {
			// use the entire match as the output value
			value = txt[m[0]:m[1]]
		}
		outMsg := constructor()
		outMsg.SetParent(msg)
		outMsg.SetPosition(msg.Position())
		outMsg.SetText(value)

		res := Result{
			Matched:  true,
			Messages: []Message{outMsg},
			Consumed: 1,
		}
		return res
	}
}

/*
// Token implements Message.  It is used to split an input message into
// multiple output messages, each representing a token.
type Token struct {
	DefaultMessage
	// regex rules for tokenizing input
	Rules []Rule
}

// Run tokenizes each input text
func (p *Token) Run(input <-chan string) <-chan Token {
	out := make(chan Token)
	go func() {
		// iterate over each input text
		inputNum := 0
		for s := range input {
			inputNum++
			// keep track of line and column number within each input text
			line := 1
			column := 1
			var value []byte
			// iterate over each byte in the input text
			for _, b := range []byte(s) {
				// accumulate the current token value
				value = append(value, b)
				// match the current token value against each rule --
				// first match wins
				for _, rule := range p.Rules {
					re := regexp.MustCompile(rule.Pattern) // XXX compile once
					m := re.FindSubmatchIndex(value)

					if m != nil && m[1] == len(value) {
						// XXX add support for submatches
						// emit the current token
						out <- Token{
							Type:    rule.Type,
							Value:   string(value),
							TextNum: inputNum,
							Line:    line,
							Column:  column,
						}
						// reset the current token value
						value = nil
						// move the column number to the end of the token
						column += len(value)
						break
					}
				}
				// if the byte is a newline, increment the line number
				// and reset the column number
				if b == '\n' {
					line++
					column = 1
				}
			}
			// if there is a token value left, emit it as an unknown token
			if len(value) > 0 {
				out <- Token{
					Type:    Unknown,
					Value:   string(value),
					TextNum: inputNum,
					Line:    line,
					Column:  column,
				}
			}
			// emit an EOF token
			out <- Token{
				Type:    EOF,
				Value:   "",
				TextNum: inputNum,
				Line:    line,
				Column:  column,
			}
		}
	}()
	return out
}
*/
