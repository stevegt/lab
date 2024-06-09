package main

import (
	"fmt"
	"os"
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
