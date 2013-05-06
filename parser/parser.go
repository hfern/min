package parser

import ()

func (p *VMTree) ParseTree() *Node {
	tree := p.TokenTree.(*tokens16)

	// unpack from channel to sliding frame
	tchan, _ := tree.PreOrder()
	pool := tokenpool{}
	pool.Enumerate(tchan, true) // skip spaces

	root := newNode()
	root.Tok.next = -1 // base tokens start at 0
	_recursiveBranch(&root, &pool, &p.Buffer, 0)

	p.ASTTree = root
	return &p.ASTTree
}

func _recursiveBranch(root *Node, pool *tokenpool, sourcecode *string, level int) {
	level++
	logRecursionHead(level, root)
	for tok := pool.Next(); tok != nil; tok = pool.Next() {
		logParseRecursion(tok)
		doRecursionPause()

		if tok.next > root.Tok.next {
			logHitChild(pool, root.Tok.next, tok.next)
			child := newNodeT(*tok)
			child.bindSource(sourcecode)
			root.addChild(&child)
			_recursiveBranch(&child, pool, sourcecode, level)
		} else {
			//logHitElse(pool)
			// Oops. A token that we shouldn't have
			// consumed
			pool.Back()
			return
		}
	}
}
