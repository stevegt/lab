package main

import (
	"syscall/js"
	// . "github.com/stevegt/goadapt"
)

func main() {

	document := js.Global().Get("document")
	p := document.Call("createElement", "p")
	p.Set("innerHTML", "Hello from Go 1.21.3")
	document.Get("body").Call("appendChild", p)
}
