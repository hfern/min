package parser

import (
	"errors"
	"fmt"
)

var SEARCH_END error = errors.New("End of Search.")

type Node struct {
	Tok      State16
	Children []*Node
	parent   *Node
	source   string
}

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

func (root *Node) EachNode(list *[]*Node, fn func(*[]*Node, *Node) bool) {
	defer func() {
		if r := recover(); r != nil {
			// Only recover from SEARCH_END
			// Do not catch other panics
			if r != SEARCH_END {
				panic(r)
			}
		}
	}()
	root._executeEach(list, fn)
	return
}

// Runs the user function on each node. To end the search, 
// call panic(parser.SEARCH_END)
func (root *Node) _executeEach(list *[]*Node, fn func(*[]*Node, *Node) bool) {
	for _, child := range root.Children {
		returncode := fn(list, child)
		if returncode {
			return
		}
		child._executeEach(list, fn)
	}
}

func (root *Node) GetNodesByRule(r Rule, onlyone bool) []*Node {
	results := make([]*Node, 0, 5)
	root.EachNode(&results, func(list *[]*Node, n *Node) bool {
		if n.Tok.Rule == r {
			*list = append(*list, n)
			if onlyone {
				panic(SEARCH_END)
			}
		}
		return false
	})
	return results
}

func (root *Node) Source() string {
	return root.source
}

func (root *Node) bindSource(s *string) {
	root.source = (*s)[root.Tok.begin:root.Tok.end]
}
