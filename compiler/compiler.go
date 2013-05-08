package compiler

import (
	//"fmt"
	"github.com/hfern/min/parser"
	//"log"
)

type Compiler struct {
	tree    *parser.VMTree
	program *Program
}

func NewCompiler() *Compiler {
	cmp := Compiler{}
	program := NewProgram()
	cmp.program = &program
	cmp.program.__compiler = &cmp
	return &cmp
}

func (c *Compiler) SetTree(t *parser.VMTree) {
	c.tree = t
}

func (c *Compiler) Compile() {
	routineNodes := c.tree.ASTTree.GetNodesByRule(parser.Ruleroutine)
	for _, node := range routineNodes {
		r := NewRoutineNP(node, c.program)
		r.lex()
	}
}
