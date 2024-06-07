package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/afero"

	"github.com/gorilla/websocket"
	. "github.com/stevegt/goadapt"
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

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: grid {subcommand} [args...]")
		fmt.Println("       grid --show {subcommand}")
		os.Exit(1)
	}

	sys := NewSys(afero.NewOsFs(), os.Getenv("HOME"))

	sys.loadPeers()
	connectToPeers()

	switch args[1] {
	case "--show":
		if len(args) < 3 {
			fmt.Println("Usage: grid --show {subcommand}")
			os.Exit(1)
		}
		subcommand := args[2]
		showPromise(sys, subcommand)
	case "start-server":
		sys.startWebSocketServer()
	default:
		subcommand := args[1]
		executeSubcommand(sys, subcommand, args[2:])
	}
}

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

func showPromise(sys *NativeSystem, subcommand string) {
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

func executeSubcommand(sys *NativeSystem, subcommand string, args []string) {
	symbolTableHash, err := sys.getSymbolTableHash()
	Ck(err)
	symbolTable := fetchSymbolTable(symbolTableHash)
	subcommandHash := getSubcommandHash(symbolTable, subcommand)
	module := sys.fetchModule(subcommandHash)
	cmd := exec.Command(module, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Println("Error executing subcommand:", err)
		os.Exit(1)
	}
}

func (sys *NativeSystem) fetchLocalData(mBuf []byte) ([]byte, error) {
	fn := fmt.Sprintf("%x", mBuf)
	cachePath := filepath.Join(sys.baseDir, cacheDir, fn)
	data, err := sys.util.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}

	// XXX If data not found in cache, check if it's a known handler
	// handlerPath := filepath.Join(os.Getenv("HOME"), gridDir, "handlers", hash)
	// return ioutil.ReadFile(handlerPath)

	return nil, fmt.Errorf("Data not found.")
}
