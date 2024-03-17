package multistage

import (
	"strings"
	"testing"

	. "github.com/stevegt/goadapt"
)

// toUpper is a stage that converts strings to upper case.
func toUpper(input <-chan string) (output <-chan string) {
	out := make(chan string, 999)
	go func() {
		for s := range input {
			out <- strings.ToUpper(s)
		}
		close(out)
	}()
	return out
}

// toQuestion is a stage that replaces '!' with '?'.
func toQuestion(input <-chan string) (output <-chan string) {
	out := make(chan string, 999)
	go func() {
		for s := range input {
			out <- strings.Replace(s, "!", "?", -1)
		}
		close(out)
	}()
	return out
}

func TestHelloWorld(t *testing.T) {
	input := "Hello, World!"

	in := make(chan string, 999)
	up := toUpper(in)
	q := toQuestion(up)
	in <- input
	close(in)
	out := ""
	for s := range q {
		out += s
	}
	Tassert(t, out == "HELLO, WORLD?", "Expected 'HELLO, WORLD?', got '%s'", out)
}

// words is an input generator stage that splits a string into words.
func words(input string) (output <-chan string) {
	out := make(chan string, 999)
	go func() {
		for _, word := range strings.Fields(input) {
			out <- word
		}
		close(out)
	}()
	return out
}

// backtrack is a stage that uses the Backtracker type.  It repeats
// the last sentence of the input string, followed by the first
// sentence.
func backtrack(input chan string) (output chan string) {
	out := make(chan string, 999)
	backtracker := NewBacktracker(input)
	go func() {
		cp1 := backtracker.Checkpoint()
		var cp2 checkpoint
		// pass through the entire input
		for _, word := range backtracker.Next() {
			out <- word
			if strings.HasSuffix(word, ".") {
				cp2 = backtracker.Checkpoint()
			}
		}
		// repeat the last sentence
		backtracker.Rollback(cp2)
		for _, word := range backtracker.Next() {
			out <- word
		}
		// repeat the first sentence
		backtracker.Rollback(cp1)
		for _, word := range backtracker.Next() {
			out <- word
			if strings.HasSuffix(word, ".") {
				break
			}
		}
		close(out)
	}()
	return out
}

func TestBackTrack(t *testing.T) {
	input := "This is a test of backtracking.  This is only a test."
	in := make(chan string, 999)
	w := words(in)
	back := backtracker(words)
	in <- input
	close(in)
	var outParts []string
	for word := range back {
		outParts = append(outParts, word)
	}
	out := strings.Join(outParts, " ")
	expect := "This is a test of backtracking. This is only a test. This is only a test. This is a test of backtracking."
	Tassert(t, out == expect, "Expected '%s', got '%s'", expect, out)
}
