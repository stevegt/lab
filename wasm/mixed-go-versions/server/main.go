package main

import (
	"net/http"

	"github.com/stevegt/envi"
	. "github.com/stevegt/goadapt"
)

func main() {
	serve()
}

var port = envi.Int("PORT", 2206)

// serve the current directory on an HTTP server.
func serve() {
	http.Handle("/", http.FileServer(http.Dir(".")))
	Pf("server listening on port %d\n", port)
	http.ListenAndServe(Spf(":%d", port), nil)
}
