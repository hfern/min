package compiler

import (
	"fmt"
	"github.com/hfern/min/parser"
)

type Compiler struct {
	routines map[string]Routine
	tree     *parser.VMTree
	program  Program
}

func NewCompiler() Compiler {
	cmp := Compiler{}
	cmp.routines = make(map[string]Routine)
	cmp.program = NewProgram()
	cmp.program.__compiler = &cmp
	return cmp
}

func (c *Compiler) SetTree(t *parser.VMTree) {
	c.tree = t
}

func (c *Compiler) Compile() {
	routineNodes := c.tree.ASTTree.GetNodesByRule(parser.Ruleroutine, false)
	for _, node := range routineNodes {

		fmt.Println(node.ChildIndex())
	}
}
