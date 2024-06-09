package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	// . "github.com/stevegt/goadapt"
)

const (
	gridDir    = ".grid"
	configFile = ".grid/config"
	cacheDir   = ".grid/cache"
	peerList   = ".grid/peers"
)

type Peer struct {
	Address string
	Conn    *websocket.Conn
}

var Peers = make(map[string]*Peer)
var mu sync.Mutex

func getSubcommandHash(symbolTable, subcommand string) string {
	lines := strings.Split(symbolTable, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, subcommand+" ") {
			return strings.TrimPrefix(line, subcommand+" ")
		}
	}
	fmt.Printf("Subcommand %s not found in symbol table.\n", subcommand)
	os.Exit(1)
	return ""
}

func (sys *KernelNative) showPromise(subcommand string) {
	symbolTableHash, err := sys.getSymbolTableHash()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	symbolTable := fetchSymbolTable(symbolTableHash)
	subcommandHash := getSubcommandHash(symbolTable, subcommand)
	module := sys.fetchModule(subcommandHash)
	cmd := exec.Command(module, "--show-promise")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing subcommand:", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}
