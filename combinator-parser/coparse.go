package coparse

import (
	"context"
	// . "github.com/stevegt/goadapt"
)

// Node interface represents the common behaviors of all nodes within the AST.
type Node interface {
	Clone() Node
	Parse(ctx context.Context) bool
	AddChildren(...Node)
	Error() string
}

// Scanner embeds position tracking functionalities within a node.
type Scanner struct {
	Start, End, Line, Column int
	// Reference to the input text
	src []byte
}

// String returns the string representation of the text spanned by the scanner.
func (s Scanner) String() string {
	return string(s.src[s.Start:s.End])
}

// CloneScanner creates a deep copy of the Scanner.
func (s Scanner) CloneScanner() Scanner {
	return Scanner{
		Start:  s.Start,
		End:    s.End,
		Line:   s.Line,
		Column: s.Column,
		src:    s.src,
	}
}

// NodeBase is an embeddable concrete node.
type NodeBase struct {
	Scanner
	// error message
	msg      string
	children []Node
}

// FromBytes initializes the node's scanner from a byte slice.
func (n *NodeBase) FromBytes(buf []byte) {
	n.Scanner = Scanner{src: buf}
}

// Clone creates a deep copy of NodeBase, fulfilling the Node interface requirement.
func (n *NodeBase) Clone() *NodeBase {
	return &NodeBase{
		Scanner: n.CloneScanner(),
		// Clone other specific fields
	}
}

// Error returns the error message of the node, fulfilling the Node interface requirement.
func (n *NodeBase) Error() string {
	return n.msg
}

// Parse() modifies the given node and returns a boolean indicating
// success or failure. If parsing fails, calling Error() on the node
// will return the error message. The Parse() method is called by
// Try() or by combinator functions to execute parsing logic.
func (n *NodeBase) Parse(ctx context.Context) bool {
	// Parsing logic that might involve calling out to AI or human agents
	// Example pseudocode:
	// if needsInsight(n) { // Assume needsInsight checks if AI/human insight is needed
	//     insight, err := requestInsightFromAgent(ctx, n)
	//     if err != nil {
	//         return nil, err
	//     }
	//     // Process insight to modify/expand the node
	// }
	// Normal parsing process follows
	n.msg = "Parse() not implemented"
	return false
}

// AddChildren adds zero or more child nodes to the node.
func (n *NodeBase) AddChildren(children ...Node) {
	n.children = append(n.children, children...)
}

// Try clones the given node, calls Parse() on the clone, and returns
// the clone and a boolean indicating success or failure
func Try(ctx context.Context, node Node) (parsedNode Node, ok bool) {
	n := node.Clone()
	ok = n.Parse(ctx)
	return n, ok
}

// A combinator is an implementation of Node.  Its constructor
// composes a new node with zero or more children.  Its Parse() method
// calls Parse() on the children per the combinator's logic and
// returns a boolean indicating success or failure.

type AndT struct {
	NodeBase
}

// And is a combinator.  It returns a new node whose children must
// parse successfully in sequence. If any child fails, the new node
// fails. If all children parse successfully, the new node succeeds.
func And(children ...Node) (n *AndT) {
	n = &AndT{}
	n.AddChildren(children...)
	return n
}

func (n *AndT) Parse(ctx context.Context) bool {
	for _, child := range n.children {
		if child == nil {
			continue
		}
		if !child.Parse(ctx) {
			n.msg = child.Error()
			return false
		}
	}
	return true
}
