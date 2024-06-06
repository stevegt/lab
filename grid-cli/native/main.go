package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

const (
	gridDir    = ".grid"
	configFile = ".grid/config"
	cacheDir   = ".grid/cache"
	peersDir   = ".grid/peers"
	peerList   = ".grid/peers/peers.list"
)

type Peer struct {
	Address string
	Conn    *websocket.Conn
}

var peers = make(map[string]*Peer)
var mu sync.Mutex

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: grid {subcommand} [args...]")
		fmt.Println("       grid --show {subcommand}")
		os.Exit(1)
	}

	ensureDirectories()
	loadPeers()
	connectToPeers()

	switch args[1] {
	case "--show":
		if len(args) < 3 {
			fmt.Println("Usage: grid --show {subcommand}")
			os.Exit(1)
		}
		subcommand := args[2]
		showPromise(subcommand)
	case "start-server":
		startWebSocketServer()
	default:
		subcommand := args[1]
		executeSubcommand(subcommand, args[2:])
	}
}

func ensureDirectories() {
	directories := []string{gridDir, cacheDir, peersDir}
	for _, dir := range directories {
		if _, err := os.Stat(filepath.Join(os.Getenv("HOME"), dir)); os.IsNotExist(err) {
			os.MkdirAll(filepath.Join(os.Getenv("HOME"), dir), os.ModePerm)
		}
	}
}

func getSymbolTableHash() string {
	configPath := filepath.Join(os.Getenv("HOME"), configFile)
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Println("Configuration file not found.")
		os.Exit(1)
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "symbol_table_hash=") {
			return strings.TrimPrefix(line, "symbol_table_hash=")
		}
	}
	fmt.Println("Symbol table hash not found in configuration.")
	os.Exit(1)
	return ""
}

func loadPeers() {
	peersPath := filepath.Join(os.Getenv("HOME"), peerList)
	file, err := os.Open(peersPath)
	if err != nil {
		fmt.Println("No peers available.")
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		peerAddress := scanner.Text()
		peers[peerAddress] = &Peer{Address: peerAddress}
	}
}

func connectToPeers() {
	var wg sync.WaitGroup
	for _, peer := range peers {
		wg.Add(1)
		go func(peer *Peer) {
			defer wg.Done()
			connectToPeer(peer)
		}(peer)
	}
	wg.Wait()
}

func connectToPeer(peer *Peer) {
	conn, _, err := websocket.DefaultDialer.Dial(peer.Address, nil)
	if err != nil {
		fmt.Printf("Failed to connect to peer %s: %v\n", peer.Address, err)
		return
	}
	mu.Lock()
	peer.Conn = conn
	mu.Unlock()
}

func queryPeers(hash, promise string) string {
	query := map[string]string{"hash": hash, "promise": promise}
	queryJSON, _ := json.Marshal(query)

	for _, peer := range peers {
		if peer.Conn == nil {
			continue
		}
		err := peer.Conn.WriteMessage(websocket.TextMessage, queryJSON)
		if err != nil {
			fmt.Printf("Failed to write to peer %s: %v\n", peer.Address, err)
			continue
		}

		_, message, err := peer.Conn.ReadMessage()
		if err != nil {
			fmt.Printf("Failed to read from peer %s: %v\n", peer.Address, err)
			continue
		}
		return string(message)
	}

	fmt.Println("Failed to fetch data from peers.")
	os.Exit(1)
	return ""
}

func fetchSymbolTable(hash string) string {
	return queryPeers(hash, "I promise to use the symbol table responsibly.")
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

func fetchModule(hash string) string {
	cachePath := filepath.Join(os.Getenv("HOME"), cacheDir, hash)
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		data := queryPeers(hash, "I promise to use this module responsibly.")
		ioutil.WriteFile(cachePath, []byte(data), 0755)
	}
	return cachePath
}

func showPromise(subcommand string) {
	symbolTableHash := getSymbolTableHash()
	symbolTable := fetchSymbolTable(symbolTableHash)
	subcommandHash := getSubcommandHash(symbolTable, subcommand)
	module := fetchModule(subcommandHash)
	cmd := exec.Command(module, "--show-promise")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Error executing subcommand:", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}

func executeSubcommand(subcommand string, args []string) {
	symbolTableHash := getSymbolTableHash()
	symbolTable := fetchSymbolTable(symbolTableHash)
	subcommandHash := getSubcommandHash(symbolTable, subcommand)
	module := fetchModule(subcommandHash)
	cmd := exec.Command(module, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		fmt.Println("Error executing subcommand:", err)
		os.Exit(1)
	}
}

func startWebSocketServer() {
	http.HandleFunc("/ws", handleWebSocket)
	fmt.Println("Starting WebSocket server on :8080")
	http.ListenAndServe(":8080", nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Failed to upgrade to websocket:", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("Failed to read message:", err)
			break
		}

		var query map[string]string
		if err := json.Unmarshal(message, &query); err != nil {
			fmt.Println("Failed to unmarshal query:", err)
			continue
		}

		hash := query["hash"]
		// promise := query["promise"]

		// Check if the requested hash is for a module or handler
		data, err := fetchLocalData(hash)
		if err != nil {
			fmt.Println("Failed to read data:", err)
			continue
		}

		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			fmt.Println("Failed to write message:", err)
			break
		}
	}
}

func fetchLocalData(hash string) ([]byte, error) {
	cachePath := filepath.Join(os.Getenv("HOME"), cacheDir, hash)
	data, err := ioutil.ReadFile(cachePath)
	if err == nil {
		return data, nil
	}

	// If data not found in cache, check if it's a known handler
	handlerPath := filepath.Join(os.Getenv("HOME"), gridDir, "handlers", hash)
	return ioutil.ReadFile(handlerPath)
}
