package compiler

import (
	"github.com/hfern/min/parser"
)

type Routine struct {
	registers RegisterMap
	vmap      VariablePool
	args      []string
	__program *Program
	__node    *parser.Node
	__name    string
}

/**
 * Lex the routine tokens for routine name, 
 * arguments, variable positions.
 */
func (r *Routine) lex() {
	r.__name = r.lex_name()
	r.register_arguments(r.lex_arguments())
	r.register_variable_positions(r.lex_variables())

}

// Determines variable positions
func (r *Routine) lex_variables() parser.NodeArray {
	codeblock := r.__node.Child(parser.Rulecodeblock)
	variables := codeblock.GetNodesByRule(parser.Rulevariable)
	// Variables that are in a parser.Rulefuncidentifier
	// are not variables but actually function call
	// function identifiers
	number_pruned := variables.Filter(func(node *parser.Node) bool {
		if parent := node.Parent(); parent != nil {
			if parent.Tok.Rule == parser.Rulefuncidentifier {
				log_function_call_saved(r.__name, node)
				return false
			}
		}
		log_variable_trace(r.__name, node)
		return true
	})
	log_number_funccalls_saved(r.__name, number_pruned)
	return variables
}

func (r *Routine) register_variable_positions(nodes parser.NodeArray) {
	for _, node := range nodes {
		r.vmap.AddInstance(node)
	}
}

/**
 * Returns name of routine from tokens
 */
func (r *Routine) lex_name() string {
	return r.__node.
		GetNodeByRule(parser.RulefuncIdDecl).
		Child(parser.Rulevariable).
		Source()
}

/**
 * Returns array of argument variable nodes to the 
 * routine in order from first argument to last
 */

func (r *Routine) lex_arguments() []*parser.Node {
	parameter_list := r.__node.
		GetNodeByRule(parser.Ruleparamaterdecl).
		GetNodeByRule(parser.Ruleparameters)

	// Ruleparameters is not required for a function statement
	if parameter_list == nil {
		return parser.NodeArray{}.CastPrimitiveLit()
		// if not there, there are no parameters to the function
	}

	return parameter_list.GetNodesByRule(parser.Rulevariable)
}

func (r *Routine) register_arguments(argnodes []*parser.Node) {
	for _, argument := range argnodes {
		r.args = append(r.args, argument.Source())
		r.vmap.AddInstance(argument)
	}
}

func NewRoutine() *Routine {
	rout := Routine{}
	rout.registers = NewRegisterMap()
	rout.vmap = NewVariablePool()
	return &rout
}

func NewRoutineN(nd *parser.Node) *Routine {
	r := NewRoutine()
	r.__node = nd
	return r
}

func NewRoutineNP(nd *parser.Node, prog *Program) *Routine {
	rout := NewRoutineN(nd)
	rout.__program = prog
	return rout
}
