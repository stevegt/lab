package v2

import (
	"context"
	"errors"
)

type Message struct {
	Parms   []interface{}          `json:"parms"`   // Parameters, with promise as the first element
	Payload map[string]interface{} `json:"payload"` // Meta information or additional data
}

type SyscallNode struct {
	modules  []Module
	children map[string]*SyscallNode
}

type Kernel struct {
	root          *SyscallNode
	knownMessages map[string]Message
}

// Module is roughly equivalent to an application in a microkernel
// system.
type Module interface {
	// HandleMessage handles an incoming message, returning a promise.
	// If `test` is true, it does not execute the actual handling
	// logic, but instead checks if the handler can handle the
	// message, and returns a promise of handling ability.
	HandleMessage(ctx context.Context, test bool, in *Message) (out *Message, err error)
}

type LocalCacheModule struct {
	cacheDir string
}

func NewLocalCacheModule(cacheDir string) *LocalCacheModule {
	return &LocalCacheModule{cacheDir: cacheDir}
}

func (m *LocalCacheModule) HandleMessage(ctx context.Context, test bool, in *Message) (out *Message, err error) {
	if test {
		// Logic to check if the module can handle the message, returning a promise of handling ability
		// XXX return a real promise
		accept := &Message{
			Parms: []interface{}{true},
			Payload: map[string]interface{}{
				"message": "Acceptable",
			},
		}
		return accept, nil
	}

	// Implement logic to handle messages
	// XXX return a real promise
	data := &Message{
		Parms: []interface{}{true},
		Payload: map[string]interface{}{
			"message": "Handled",
		},
	}
	return data, nil
}

func (k *Kernel) consultModules(ctx context.Context, in *Message) ([]byte, error) {
	node := k.root
	for _, parm := range in.Parms {
		if child, ok := node.children[parm.(string)]; ok {
			node = child
		} else {
			break
		}
	}

	var promisingModules []Module
	for _, module := range node.modules {
		promise, err := module.HandleMessage(ctx, true, in)
		if err != nil || !promise.Parms[0].(bool) {
			continue
		}
		promisingModules = append(promisingModules, module)
		k.addSyscall(in.Parms...)
	}

	for _, module := range promisingModules {
		response, err := module.HandleMessage(ctx, false, in)
		if err == nil && response.Parms[0].(bool) {
			return []byte("Handled"), nil
		}
	}

	return nil, errors.New("no module could handle the request")
}

func (k *Kernel) addSyscall(parms ...interface{}) {
	node := k.root
	for _, parm := range parms {
		strParm := parm.(string)
		if _, ok := node.children[strParm]; !ok {
			node.children[strParm] = &SyscallNode{
				children: make(map[string]*SyscallNode),
			}
		}
		node = node.children[strParm]
	}
	node.modules = append(node.modules, k.knownMessages[parms[0].(string)])
}
