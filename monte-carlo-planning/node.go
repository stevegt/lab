package mcats

// Node represents a node in the MCaTS analysis.
type Node struct {
	Name     string   // Name is a unique identifier for the node.
	Desc     string   // Desc describes the purpose or action of the node.
	Duration int      // Duration is the time taken to complete the node's action.
	Preqs    []string // Preqs holds names of prerequisite nodes.
}

// NodeSet represents a set of nodes in an MCaTS analysis.
type NodeSet struct {
	Nodes map[string]*Node // Nodes maps Node names to Node structures.
}

// NewNodeSet creates a new NodeSet with the given nodes.
// It also ensures that duplicate nodes are handled correctly.
func NewNodeSet(nodes ...*Node) *NodeSet {
	ns := &NodeSet{Nodes: make(map[string]*Node)}
	for _, node := range nodes {
		ns.Nodes[node.Name] = node
	}
	return ns
}

// Verify checks the prerequisites for each node in a NodeSet are valid
// and not self-referential or circular.
func (ns *NodeSet) Verify() bool {
	if len(ns.Nodes) == 0 {
		return true // An empty set passes verification by default
	}

	// Check for self-referential prerequisites
	for _, node := range ns.Nodes {
		for _, preq := range node.Preqs {
			if preq == node.Name {
				return false // Found self-referential prerequisite
			}
		}
	}

	// Helper function to recursively check for circular dependencies.
	var checkCircular func(string, map[string]bool) bool
	checkCircular = func(nodeName string, visited map[string]bool) bool {
		if visited[nodeName] {
			return true // Circular dependency detected
		}
		if _, exists := ns.Nodes[nodeName]; !exists {
			return true // Prerequisite does not exist
		}

		visited[nodeName] = true
		for _, preq := range ns.Nodes[nodeName].Preqs {
			if checkCircular(preq, visited) {
				return true
			}
		}
		delete(visited, nodeName) // Remove from visited as we move up the stack
		return false
	}

	// Check for missing or circular prerequisites.
	for nodeName := range ns.Nodes {
		if checkCircular(nodeName, make(map[string]bool)) {
			return false // Circular or missing prerequisites detected
		}
	}

	return true
}
