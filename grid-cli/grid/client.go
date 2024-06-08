package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/gorilla/websocket"
)

func connectToPeers() {
	var wg sync.WaitGroup
	for _, peer := range Peers {
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

	for _, peer := range Peers {
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

func (sys *NativeKernel) fetchModule(hash string) string {
	cachePath := filepath.Join(sys.baseDir, cacheDir, hash)
	if _, err := sys.fs.Stat(cachePath); os.IsNotExist(err) {
		data := queryPeers(hash, "I promise to use this module responsibly.")
		ioutil.WriteFile(cachePath, []byte(data), 0755)
	}
	return cachePath
}
