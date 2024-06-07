package main

import (
	"bufio"
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

// Sys represents the underlying system.
type Sys struct {
	Fs      afero.Fs
	BaseDir string
	util    *afero.Afero
}

// NewSys creates a new Sys.
func NewSys(fs afero.Fs, baseDir string) *Sys {
	sys := &Sys{
		Fs:      fs,
		BaseDir: baseDir,
		util:    &afero.Afero{Fs: fs},
	}
	sys.ensureDirectories()
	return sys
}

func (sys *Sys) ensureDirectories() {
	directories := []string{gridDir, cacheDir}
	for _, dir := range directories {
		if _, err := sys.Fs.Stat(filepath.Join(sys.BaseDir, dir)); os.IsNotExist(err) {
			sys.Fs.MkdirAll(filepath.Join(sys.BaseDir, dir), os.ModePerm)
		}
	}
}

func (sys *Sys) getSymbolTableHash() (hash string, err error) {
	configPath := filepath.Join(sys.BaseDir, configFile)
	data, err := sys.util.ReadFile(configPath)
	if err != nil {
		err = fmt.Errorf("Failed to read configuration: %v", err)
		return "", err
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "symbol_table_hash=") {
			return strings.TrimPrefix(line, "symbol_table_hash="), nil
		}
	}
	err = fmt.Errorf("Symbol table hash not found in configuration.")
	return "", err
}

func (sys *Sys) loadPeers() {
	peersPath := filepath.Join(sys.BaseDir, peerList)
	file, err := sys.Fs.Open(peersPath)
	if err != nil {
		fmt.Println("No peers available.")
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		peerAddress := scanner.Text()
		Peers[peerAddress] = &Peer{Address: peerAddress}
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

func showPromise(sys *Sys, subcommand string) {
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

func executeSubcommand(sys *Sys, subcommand string, args []string) {
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

func (sys *Sys) fetchLocalData(mBuf []byte) ([]byte, error) {
	fn := fmt.Sprintf("%x", mBuf)
	cachePath := filepath.Join(sys.BaseDir, cacheDir, fn)
	data, err := sys.util.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}

	// XXX If data not found in cache, check if it's a known handler
	// handlerPath := filepath.Join(os.Getenv("HOME"), gridDir, "handlers", hash)
	// return ioutil.ReadFile(handlerPath)

	return nil, fmt.Errorf("Data not found.")
}
