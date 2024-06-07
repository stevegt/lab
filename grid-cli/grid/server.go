package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	. "github.com/stevegt/goadapt"
)

func (sys *Sys) startWebSocketServer() {
	http.HandleFunc("/ws", sys.handleWebSocket)
	fmt.Println("Starting WebSocket server on :8080")
	http.ListenAndServe(":8080", nil)
}

func (sys *Sys) handleWebSocket(w http.ResponseWriter, r *http.Request) {
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

		mStr := query["hash"]
		// convert multihash hex string to byte slice
		mBuf, err := hex.DecodeString(mStr)
		Ck(err)

		// promise := query["promise"]

		// Check if the requested hash is for a module or handler
		data, err := sys.fetchLocalData(mBuf)
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
