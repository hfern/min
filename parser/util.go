package parser

func newNode() Node {
	return Node{Children: make([]*Node, 0, 1)}
}

func newNodeT(t State16) Node {
	n := newNode()
	n.Tok = t
	return n
}

func newNodeTP(state State16, root *Node) Node {
	child := newNodeT(state)
	root.addChild(&child)
	return child
}

func whitespacefilter(in <-chan State16) <-chan State16 {
	out := make(chan State16, 0)
	go func() {
		for state := range in {
			if !is_whitespace(state.Rule) {
				out <- state
			}
		}

		close(out)
	}()
	return out
}

func is_whitespace(r Rule) bool {
	return r == Rulespace ||
		r == Ruleoptspace ||
		r == Ruleminspace ||
		r == Ruleliteralspace ||
		r == Rulecommentdoubleslash ||
		r == Rulecomment ||
		r == Rulecommentblock
}

func get_n_spaces(n int) string {
	if n == 1 {
		return " "
	}
	b := make([]byte, n, n)
	for i := 0; i < n; i++ {
		b[i] = 32 // ASCII dec 32 == space
	}
	return string(b)
}
