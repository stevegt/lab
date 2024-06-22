package main

import (
	"fmt"
	"os"

	"github.com/spf13/afero"
	. "github.com/stevegt/goadapt"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: grid {subcommand} [args...]")
		fmt.Println("       grid --show {subcommand}")
		os.Exit(1)
	}

	sys := NewKernelNative(afero.NewOsFs(), os.Getenv("HOME"))

	sys.loadPeers()
	connectToPeers()

	switch args[1] {
	case "--show":
		if len(args) < 3 {
			fmt.Println("Usage: grid --show {subcommand}")
			os.Exit(1)
		}
		subcommand := args[2]
		sys.showPromise(subcommand)
	case "start-server":
		sys.startWebSocketServer()
	default:
		subcommand := args[1]
		err := sys.Exec(subcommand, args[2:])
		Ck(err)
	}
}
