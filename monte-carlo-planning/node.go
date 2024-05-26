package mcats

// Node represents a node in the monte-carlo tree search.
type Node struct {
	Name     string   // Unique identifier for the node.
	Desc     string   // Description of the node.
	Duration int      // Duration taken by the node.
	Preqs    []string // Prerequisites of the node.
}

// NodeSet represents a set of nodes and their order. It manages the addition and verification of nodes.
type NodeSet struct {
	Nodes map[string]*Node
	Order []string // Tracks the order of nodes to validate prerequisites accurately.
}

// NewNodeSet creates and returns a new instance of NodeSet initialized with the provided nodes.
// It also ensures that there are no duplicate nodes and maintains their order.
func NewNodeSet(nodes ...*Node) *NodeSet {
	ns := &NodeSet{
		Nodes: make(map[string]*Node),
		Order: make([]string, 0, len(nodes)),
	}
	for _, node := range nodes {
		if _, exists := ns.Nodes[node.Name]; !exists {
			ns.Order = append(ns.Order, node.Name)
		}
		ns.Nodes[node.Name] = node
	}
	return ns
}

// Verify checks the prerequisites of each node in the set to ensure they are met within the set in the correct order.
// It also checks for self-referencing, circular dependencies, and missing prerequisites.
func (ns *NodeSet) Verify() bool {
	seen := make(map[string]bool)
	for _, nodeName := range ns.Order {
		node, exists := ns.Nodes[nodeName]
		if !exists { // This should never happen.
			return false
		}
		for _, preq := range node.Preqs {
			// Check for self-referential prerequisite.
			if preq == node.Name {
				return false
			}
			// Check if the prerequisite exists and it has been seen (ordered before).
			if _, exists := ns.Nodes[preq]; !exists || !seen[preq] {
				return false
			}
		}
		seen[nodeName] = true
	}
	return true
}

