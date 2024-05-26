package mcats

// Node defines a single node in the search space with its basic attributes.
type Node struct {
	Name     string   // the name of the node
	Desc     string   // description of the node
	Duration int      // the duration or cost of the node
	Preqs    []string // the prerequisites of the node
}

// NodeSet represents a set of nodes.
type NodeSet struct {
	Nodes map[string]*Node
}

// NewNodeSet creates a new NodeSet from a variadic list of nodes.
func NewNodeSet(nodes ...*Node) *NodeSet {
	ns := &NodeSet{
		Nodes: make(map[string]*Node),
	}
	for _, node := range nodes {
		ns.Nodes[node.Name] = node
	}
	return ns
}

// Verify checks if the prerequisites of all nodes in the set are satisfied within the set itself.
func (ns *NodeSet) Verify() bool {
	for _, node := range ns.Nodes {
		for _, preq := range node.Preqs {
			if _, ok := ns.Nodes[preq]; !ok {
				return false // A prerequisite is not satisfied.
			}
		}
	}
	return true // All prerequisites are satisfied.
}

