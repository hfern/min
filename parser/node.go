package parser

type Node struct {
	Tok      token32
	Children []*Node
	parent   *Node
}

func (root *Node) addChild(n *Node) {
	root.Children = append(root.Children, n)
}
