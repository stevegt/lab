package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/spf13/afero"
)

// SyscallNode represents a node in the hierarchical syscall tree
type SyscallNode struct {
	modules  []Module
	children map[string]*SyscallNode
}

// Kernel struct with the syscall tree root and file system abstraction
type Kernel struct {
	root    *SyscallNode
	fs      afero.Fs
	modules map[string]Module // Known modules
}

// NewKernel initializes a new Kernel instance with embedded modules
func NewKernel() *Kernel {
	return &Kernel{
		root: &SyscallNode{
			children: make(map[string]*SyscallNode),
		},
		fs:      afero.NewOsFs(),
		modules: make(map[string]Module), // Initialize with known modules
	}
}

// HandleWebSocket handles incoming WebSocket connections and messages
func (k *Kernel) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade to WebSocket: %v\n", err)
		return
	}
	defer conn.Close()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("Failed to read message: %v\n", err)
			break
		}
		k.processMessage(context.Background(), message)
	}
}

// processMessage processes a message, performing routing and handling using the syscall tree
func (k *Kernel) processMessage(ctx context.Context, message []byte) {
	var msg Message
	if err := json.Unmarshal(message, &msg); err != nil {
		fmt.Printf("Failed to unmarshal message: %v\n", err)
		return
	}

	// Consult modules based on the message parms
	response, err := k.consultModules(ctx, msg.Parms...)
	if err != nil {
		fmt.Printf("Failed to consult modules: %v\n", err)
		return
	}

	// Response handling (omitted for brevity)
	fmt.Println("Message processed, response generated.", response)
}

// consultModules consults modules for handling the message, utilizing the syscall tree for routing
func (k *Kernel) consultModules(ctx context.Context, parms ...interface{}) ([]byte, error) {
	// Implementation involves traversing the syscall tree to find best match and consulting modules
	// Omitted for brevity
	return nil, nil
}
