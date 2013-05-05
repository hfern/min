package parser

func newNode() Node {
	return Node{Children: make([]*Node, 0, 1)}
}

func newNodeT(t token32) Node {
	n := newNode()
	n.Tok = t
	return n
}

func newNodeTP(t token32, p *Node) Node {
	n := newNodeT(t)
	n.parent = p
	return n
}

func whitespacefilter(in <-chan token32, out chan<- token32) {
	defer close(out)
	var t token32
	var ok bool
	for {
		if t, ok = <-in; !ok {
			return
		}
		if !is_whitespace(t) {
			out <- t
		}
	}
}

func is_whitespace(tok token32) bool {
	r := tok.Rule
	return r == Rulespace ||
		r == Ruleoptspace ||
		r == Ruleminspace ||
		r == Ruleliteralspace ||
		r == Rulecommentdoubleslash ||
		r == Rulecomment ||
		r == Rulecommentblock
}
