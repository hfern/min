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

func (c *Compiler) SetSource(source string) {
	c.program.sourcecode = source
}

func (c *Compiler) Compile() error {
	var err error = nil

	defer func() {
		rec_err := recover()
		if rec_err != nil {
			err = rec_err.(error)
		}
		return
	}()

	routineNodes := c.tree.ASTTree.GetNodesByRule(parser.Ruleroutine)

	for _, node := range routineNodes {
		rout := NewRoutineNP(node, c.program)
		c.program.addRoutine(rout)
	}

	if err = c.lex_all(); err != nil {
		return err
	}

	if err = c.generate_ir(); err != nil {
		return err
	}

	return err
}

func (c *Compiler) lex_all() error {
	// TODO: Consider making function
	// process routines in paralell
	for _, rout := range c.program.routines {
		rout.lex()
		err := c.program.linkRoutine(rout)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) generate_ir() error {
	for _, rout := range c.program.routines {
		err := rout.generate_ir()
		if err != nil {
			return err
		}
	}
	return nil
}
