package mcats

import (
	"fmt"
	"testing"
	// . "github.com/stevegt/goadapt"
)

/*
name: monte-carlo tree search (MCaTS)

Similar to godecide, but instead of a decision tree in which each node
contains cost/benefit information and children with probabilities, we
have a cloud of nodes with cost/benefit/time information and
prerequisites. We randomly generate node sequences, discard those that
don't meet satisfy the prerequisite chains, and then evaluate the
cost/benefit/time of the remaining sequences. We then sort the
sequences by cost/benefit/time and present the top N to the user. The
presentation might be a table, a graph, or a timeline.

We can generate probabilities for each edge by counting the ratio of
sequences that contain the edge to the total number of sequences.

Improving on that, we could identify the most common sequence
fragments in the top N sequences and use those to generate a decision
tree that, while not necessarily optimal, would support more
contingencies.  In the process, we can identify critical sequence
fragments for which there are no alternatives, allowing the user to
focus on adding alternate paths.
*/

// example of a node set
func ExampleNodeSet() {
	// create node set
	n1 := &Node{
		Name:     "prereq1",
		Desc:     "this is a prereq node",
		Duration: 5,
	}

	n2 := &Node{
		Name:     "prereq2",
		Desc:     "this is a prereq node",
		Duration: 6,
	}

	n3 := &Node{
		Name:     "test",
		Desc:     "this is a test node",
		Duration: 10,
		Preqs:    []string{"prereq1", "prereq2"},
	}

	// create a set with the nodes in the right order
	set := NewNodeSet(n1, n2, n3)
	// verify the prereqs
	ok := set.Verify()
	fmt.Printf("nodes in the right order: ok=%t\n", ok)

	// create a node set with the nodes in the wrong order
	set = NewNodeSet(n3, n1, n2)
	// verify the prereqs
	ok = set.Verify()
	fmt.Printf("nodes in the wrong order: ok=%t\n", ok)

	// Output:
	// nodes in the right order: ok=true
	// nodes in the wrong order: ok=false
}

func TestMissingAndCircularPrerequisites(t *testing.T) {
	n1 := &Node{Name: "circle1", Preqs: []string{"circle2"}}
	n2 := &Node{Name: "circle2", Preqs: []string{"circle1"}}
	n3 := &Node{Name: "missingPre", Preqs: []string{"nonexistent"}}
	set := NewNodeSet(n1, n2, n3)
	if set.Verify() {
		t.Error("Verify() should fail on circular or missing prerequisites")
	}
}

func TestEmptyNodeSet(t *testing.T) {
	set := NewNodeSet()
	if !set.Verify() {
		t.Error("Verify() should pass for an empty NodeSet")
	}
}

func TestSelfReferentialPrerequisite(t *testing.T) {
	n1 := &Node{Name: "selfRef", Preqs: []string{"selfRef"}}
	set := NewNodeSet(n1)
	if set.Verify() {
		t.Error("Verify() should fail on self-referential prerequisites")
	}
}

func TestDuplicateNodes(t *testing.T) {
	n1 := &Node{Name: "test", Desc: "Original", Duration: 10}
	n2 := &Node{Name: "test", Desc: "Duplicate", Duration: 5}
	set := NewNodeSet(n1, n2)
	if len(set.Nodes) != 1 {
		t.Error("Duplicate nodes should not be added or should replace the original")
	}
	if set.Nodes["test"].Duration != 5 {
		t.Error("Duplicate node should update the existing node's attributes")
	}
}

func ExampleDuration() {
	// create a node set
	n1 := &Node{Name: "prereq1", Duration: 5}
	n2 := &Node{Name: "prereq2", Duration: 6}
	n3 := &Node{Name: "test", Duration: 10, Preqs: []string{"prereq1", "prereq2"}}

	// create a node set with the nodes in the right order
	set := NewNodeSet(n1, n2, n3)

	// calculate the duration of the node set
	duration := set.Duration()
	fmt.Printf("duration=%.1f\n", duration)

	// Output:
	// duration=21.0
}

func ExampleFitness() {
	// create a node set
	n1 := &Node{Name: "prereq1", Duration: 5}
	n2 := &Node{Name: "prereq2", Duration: 6}
	n3 := &Node{Name: "test", Duration: 10, Preqs: []string{"prereq1", "prereq2"}}

	// create a node set with the nodes in the right order
	set := NewNodeSet(n1, n2, n3)

	// evaluate the fitness of the node set -- fitness is 1/(sum of
	// the durations of the nodes)
	fitness := set.Fitness()
	fmt.Printf("fitness=%.3f\n", fitness)

	// Output:
	// fitness=0.048

}
