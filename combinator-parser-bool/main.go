package main

import . "github.com/stevegt/goadapt"

type Node interface {
	Expect(bool) Node
}

type NodeBase struct {
	checkpoint *NodeBase
}

func (n *NodeBase) Commit() {
	cp := *n
	n.checkpoint = &cp
}

func (n *NodeBase) Rollback() {
	*n = *n.checkpoint
}

func (n *NodeBase) Expect(cond bool) Node {
	if !cond {
		n.Rollback()
		return nil
	}
	n.Commit()
	return n
}

type NodeA struct {
	NodeBase
}

func (n *NodeA) Foo() bool {
	return true
}

func (n *NodeA) Bar() bool {
	return false
}

func main() {

	a := &NodeA{}
	b := a.Expect(a.Foo())
	c := b.Expect(b.Bar())
	d := c.Expect(c.Foo())

	Pf("a: %#v\nb: %#v\nc: %#v\nd: %#v\n", a, b, c, d)
}
