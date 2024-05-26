package mcats

import (
	"fmt"
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
func ExampleSet() {
	// create node set
	n1 := &Node{
		Name:     "test",
		Desc:     "this is a test node",
		Duration: 10,
		Preqs:    []string{"prereq1", "prereq2"},
	}

	n2 := &Node{
		Name:     "prereq1",
		Desc:     "this is a prereq node",
		Duration: 5,
	}

	n3 := &Node{
		Name:     "prereq2",
		Desc:     "this is a prereq node",
		Duration: 6,
	}

	set := NewNodeSet(n1, n2, n3)

	// verify the prereqs
	ok := set.Verify()
	fmt.Println(ok)

	// Output:
	// true
}
