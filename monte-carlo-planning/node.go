package mcats

import "errors"

// Node represents an individual task or activity in the node cloud.
type Node struct {
	Name     string   // Unique identifier for the node.
	Desc     string   // Description of what the node represents.
	Duration float64  // Time in arbitrary units that the node takes to complete.
	Preqs    []string // Prerequisite nodes that must be completed before this one.
}

// NodeSet represents a set of nodes, where each node can have prerequisites that are also in the set.
type NodeSet struct {
	Nodes map[string]*Node // Map of node names to nodes for quick lookup.
}

// NewNodeSet creates a new NodeSet containing the provided nodes.
func NewNodeSet(nodes ...*Node) *NodeSet {
	ns := &NodeSet{Nodes: make(map[string]*Node)}
	for _, node := range nodes {
		ns.Nodes[node.Name] = node // This will update the node if it’s already in the set.
	}
	return ns
}

// Verify checks the NodeSet for any missing or circular prerequisites and ensures all prerequisites are satisfied.
func (ns *NodeSet) Verify() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for _, node := range ns.Nodes {
		if !ns.dfs(node, visited, recStack) {
			return false // Cycle or missing prerequisite detected.
		}
	}
	return true
}

// dfs performs a Depth-First Search to detect cycles and ensure all prerequisites exist.
func (ns *NodeSet) dfs(node *Node, visited, recStack map[string]bool) bool {
	nodeName := node.Name
	if recStack[nodeName] {
		return false // Cycle detected
	}
	if visited[nodeName] {
		return true
	}

	visited[nodeName] = true
	recStack[nodeName] = true
	for _, preqName := range node.Preqs {
		preq, exists := ns.Nodes[preqName]
		if !exists || !ns.dfs(preq, visited, recStack) {
			return false // Missing prerequisite or cycle detected in prerequisites
		}
	}

	recStack[nodeName] = false // Backtrack
	return true
}

// Duration calculates the total duration of all nodes in the NodeSet.
func (ns *NodeSet) Duration() float64 {
	totalDuration := 0.0
	for _, node := range ns.Nodes {
		totalDuration += node.Duration
	}
	return totalDuration
}

// Fitness calculates the fitness of the NodeSet as the reciprocal of the sum of the durations.
// It returns an error if the total duration is zero to prevent division by zero.
func (ns *NodeSet) Fitness() (float64, error) {
	totalDuration := ns.Duration()
	if totalDuration == 0 {
		return 0, errors.New("total duration is zero, cannot calculate fitness")
	}
	return 1 / totalDuration, nil
}
