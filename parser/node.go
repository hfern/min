package parser

import (
	"fmt"
)

type SearchCode int
type SearchFunction func(*Node) SearchCode
type NodeArrayFilter func(*Node) bool

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

func (nd *Node) Parent() *Node {
	return nd.parent
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
func (root *Node) _performTreeSearch(r Rule, onlyone bool) NodeArray {
	results := make(NodeArray, 0, 5)

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

func (root *Node) GetNodesByRule(r Rule) NodeArray {
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

// Source returns the source code that this token was
// emitted from.
func (root *Node) Source() string {
	return root.source
}

/**
 * Child is a convenience function for getting any 
 * direct child that matches a given Rule.
 * O(n) for n = len(Children)
 * 
 * TODO: Reduce to O(log n) using hueristics on Token 
 * generation postions in source.
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

// ChildIndex attempts to determine the index
// of a given node in its parent's children array.
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

// Before returns the older sibling of the selected
// node or nil if there is none.
func (me *Node) Before() *Node {
	myindex := me.ChildIndex()
	if myindex == -1 {
		// Node is tree root; no parent
		return nil
	}
	if myindex >= 1 {
		return me.parent.Children[myindex-1]
	}
	return nil
}

// After returns the younger sibling of the selected
// node or nil if there is none.
func (me *Node) After() *Node {
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

// ToNodeArray is an alias for []NodeArray{*Node}
func (me *Node) ToNodeArray() NodeArray {
	return NodeArray{me}
}

func NewNodeArray_Cap(capacity int) NodeArray {
	return make(NodeArray, 0, capacity)
}

func NewNodeArray() NodeArray {
	return NewNodeArray_Cap(3)
}

// CastPrimitive is a convenience function that
// returns the NodeArray in the base
// []*Node format.
func (arr *NodeArray) CastPrimitive() []*Node {
	return (*arr).CastPrimitiveLit()
}

// CastPrimitiveLit is the literal variant of CastPrimitive.
// While not being especially useful, it helps cast 
// single liners in the primitive base type of NodeArray.
// 
// Example: 
//   NodeArray{}.CastPrimitiveLit()
func (arr NodeArray) CastPrimitiveLit() []*Node {
	return ([]*Node)(arr)
}

// Filter takes a function and applies it to all elements
// of the array. If the function returns true, the node
// is kept in the array. If the filter function returns 
// false, the node is removed from the array.
// 
// Returns number of nodes deleted from the array
func (arr *NodeArray) Filter(filterfunc NodeArrayFilter) int {
	newarr := NewNodeArray_Cap(len(*arr))
	disposed_nodes := 0
	for _, node := range *arr {
		if filterfunc(node) {
			newarr = append(newarr, node)
		} else {
			disposed_nodes++
		}
	}
	*arr = newarr
	return disposed_nodes
}
