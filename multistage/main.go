package main

import (
	"fmt"
	"strings"
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

func main() {
	input := "Hello, World!"

	in := make(chan string, 999)
	up := toUpper(in)
	q := toQuestion(up)
	in <- input
	close(in)
	for s := range q {
		fmt.Println(s)
	}
}
