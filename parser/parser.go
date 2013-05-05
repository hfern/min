package parser

import (
	"fmt"
)

func (p *VMTree) ParseTree() {
	ch := p.TokenTree.Tokens()
	program := newNode()
	program.Tok.next = 0 // root node

	ok, leftover := p._recursiveBranch(&program, ch)
	fmt.Println(ok)
	fmt.Println(*leftover)
}

func (p *VMTree) _recursiveBranch(root *Node, ch <-chan token32) (bool, *token32) {
	for {
		t, ok := <-ch
		if !ok {
			break
		}
	AfterTokenSelect:
		if root.Tok.isParentOf(t) {
			child := newNodeTP(t, root)
			root.addChild(&child)
			descentok, leftover := p._recursiveBranch(&child, ch)
			if !descentok {
				return descentok, leftover
			}
			if leftover != nil {
				// we have a sibling to child
				t = *leftover
				goto AfterTokenSelect
			}
		} else {
			// sibling of parent of root
			return true, &t
		}
	}
	return true, nil
}
