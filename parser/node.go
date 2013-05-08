package parser

import (
	"fmt"
)

type SearchCode int
type SearchFunction func(*Node) SearchCode

const (
	_                          = iota
	SEARCH_CONTINUE SearchCode = 1 << iota
	SEARCH_END
	SEARCH_ADD
)

type Node struct {
	Tok      State16
	Children []*Node
	parent   *Node
	source   string
}

type NodeArray []*Node

func (root *Node) addChild(n *Node) {
	root.Children = append(root.Children, n)
	n.parent = root
}

func (root *Node) recursivePrintChildren(identlevel int) {
	//
	// [Ruleprogram] =>
	//   1. [Ruleroutine] =>
	//   ...
	//
	for i, nd := range root.Children {
		fmt.Printf("%s %d. [%s] =>\n",
			get_n_spaces(identlevel),
			i,
			Rul3s[nd.Tok.Rule],
		)
		nd.recursivePrintChildren(identlevel + 1)
	}
}

func (root *Node) PrintTree() {
	fmt.Printf("[%s] => \n", Rul3s[root.Tok.Rule])
	root.recursivePrintChildren(1)
	fmt.Println(len(root.Children[1].Children))
}

func (root *Node) EachNode(fn SearchFunction) {
	//results := make([]*Node, 0, 0)
	root._executeEach(fn)
	return
}

// Runs the user function on each node. To end the search, 
// return SEARCH_END
func (root *Node) _executeEach(fn SearchFunction) SearchCode {
	var child *Node
	// If a "invalid memory address or nil pointer dereference"
	// panic is thrown referencing the following line,
	// root is a nil pointer. Check the caller of this function.
	for _, child = range root.Children {
		returncode := fn(child)
		if returncode&SEARCH_END == SEARCH_END {
			return returncode
		}
		recursion_result := child._executeEach(fn)
		if recursion_result&SEARCH_END == SEARCH_END {
			return recursion_result
		}
	}
	return SEARCH_CONTINUE
}

// TODO make this use channels for appending to 
// the results array. Removes block that would allow
// _executeEach to run on multiple goroutines
func (root *Node) _performTreeSearch(r Rule, onlyone bool) []*Node {
	results := make([]*Node, 0, 5)

	results_channel := make(chan *Node)
	done_channel := make(chan bool)

	// Syncs access to results. This allows making a 
	// multithreaded EachNode.
	go func() {
		for {
			select {
			case nd := <-results_channel:
				results = append(results, nd)
				break
			case <-done_channel:
				return
				break
			}
		}
		return
	}()

	root.EachNode(func(n *Node) SearchCode {
		if n.Tok.Rule == r {
			results_channel <- n
			if onlyone {
				return SEARCH_END
			}
		}
		return SEARCH_CONTINUE
	})

	done_channel <- true // signal end of search
	close(done_channel)
	close(results_channel)
	return results
}

func (root *Node) GetNodesByRule(r Rule) []*Node {
	return root._performTreeSearch(r, false)
}

func (root *Node) GetNodeByRule(r Rule) *Node {
	matching_nodes := root._performTreeSearch(r, true)
	if len(matching_nodes) == 0 {
		// no matches
		return nil
	}
	return matching_nodes[0]
}

func (root *Node) Source() string {
	return root.source
}

/**
 * Convenience function for getting any 
 * direct child that matches Rule
 * O(n) for n = len(Children)
 */
func (root *Node) Child(r Rule) *Node {
	for _, n := range root.Children {
		if n.Tok.Rule == r {
			return n
		}
	}
	return nil
}

func (root *Node) bindSource(s *string) {
	root.source = (*s)[root.Tok.begin:root.Tok.end]
}

func (me *Node) ChildIndex() int {
	// -1 is not found code
	myindex := -1
	if me.parent == nil {
		return myindex
	}
	for i, child := range me.parent.Children {
		if child == me {
			myindex = i
			break
		}
	}
	return myindex
}

func (me *Node) Next() *Node {
	myindex := me.ChildIndex()
	if myindex == -1 {
		// Can't find this node in parent's children.
		// Do I even exist? 
		return nil
	}

	if myindex >= len(me.parent.Children) {
		// This node has no younger sibling
		return nil
	}
	return me.parent.Children[myindex+1]
}

func (me *Node) ToNodeArray() NodeArray {
	return NodeArray{me}
}

// CastPrimitive is a convenience function that
// returns the NodeArray in the base
// []*Node format.
func (arr *NodeArray) CastPrimitive() []*Node {
	return (*arr).CastPrimitiveLit()
}

func (arr NodeArray) CastPrimitiveLit() []*Node {
	return ([]*Node)(arr)
}
