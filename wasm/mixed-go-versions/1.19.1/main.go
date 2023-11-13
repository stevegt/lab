package main

import (
	"runtime"
	"syscall/js"

	. "github.com/stevegt/goadapt"
)

func main() {
	version := runtime.Version()
	Pl("running Go %s", version)
	document := js.Global().Get("document")
	div := document.Call("getElementById", version)
	div.Set("innerHTML", Spf("Hello from %s", version))
}
