package mcats

// Node represents an individual point within a Monte Carlo analysis tree.
type Node struct {
	Name     string   // Unique identifier for the node
	Desc     string   // Descriptive text about the node's purpose or function
	Duration int      // Estimated duration that the node represents
	Preqs    []string // A list of names of nodes that are prerequisites for this node
}

// NodeSet aggregates multiple Node instances, facilitating operations upon the collective group.
type NodeSet struct {
	Nodes map[string]*Node // Key-value pair matching node names to their respective Node instances
}

// NewNodeSet provides initialization logic for a NodeSet, populating it with an initial set of Node elements.
func NewNodeSet(nodes ...*Node) *NodeSet {
	ns := &NodeSet{
		Nodes: make(map[string]*Node),
	}
	for _, node := range nodes {
		ns.Nodes[node.Name] = node
	}
	return ns
}

// Verify examines the NodeSet for any logical inconsistencies such as unmet prerequisites or circular dependencies.
func (ns *NodeSet) Verify() bool {
	// Function to recursively check for existing and correctly ordered prerequisites.
	var checkOrder func(node *Node, seen map[string]bool) bool
	checkOrder = func(node *Node, seen map[string]bool) bool {
		// Mark this node as seen for this path to help detect circular dependencies.
		if seen[node.Name] {
			return false
		}
		seen[node.Name] = true

		for _, pre := range node.Preqs {
			pNode, pExists := ns.Nodes[pre]
			if !pExists || !checkOrder(pNode, seen) {
				return false
			}
		}
		// Unmark as seen to allow revisiting through different paths.
		seen[node.Name] = false
		return true
	}

	// Verify all nodes to ensure their prerequisites are satisfied.
	for _, node := range ns.Nodes {
		if !checkOrder(node, make(map[string]bool)) {
			return false
		}
	}
	return true
}
