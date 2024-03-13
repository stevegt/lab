package main

import (
	"fmt"

	. "github.com/stevegt/goadapt"
	"github.com/tiktoken-go/tokenizer"
)

func main() {
	enc, err := tokenizer.Get(tokenizer.Cl100kBase)
	Ck(err)

	// this should print a list of token ids
	ids, _, _ := enc.Encode("this should print a list of token ids, and then the original string back; this is a test")
	fmt.Println(ids)

	// this should print the original string back
	text, _ := enc.Decode(ids)
	fmt.Println(text)
}
