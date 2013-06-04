package compiler

import (
	"github.com/hfern/min/parser"
)

func genir_codestatement(r *Routine, node *parser.Node) error {
	return handlepairs(r, node.Children[0], []ruleHandler{
		{parser.Ruleoperation, genir_operation},
		{parser.Rulelogicblock, genir_logicblock},
	})
}

func genir_operation(rout *Routine, node *parser.Node) error {
	// opaction is ALWAYS the only child
	action_node := node.Children[0].Children[0]
	return handlepairs(rout, action_node, []ruleHandler{
		{parser.Rulereservation, genir_reservation},
		{parser.Rulereturning, genir_returning},
		{parser.Ruleassignment, genir_assignment},
		{parser.Rulelabeling, genir_labeling},
		{parser.Rulejumping, genir_jumping},
	})
}

func reserve_variable(r *Routine, node *parser.Node) error {
	varname := node.Source()
	variable := r.vmap._map[varname]
	variable.Allocate()
	variable.register = r.registers.ReserveRegister()
	if variable.register == nil {
		return errorCannotReserveRegister(r, varname, node)
	}
	return nil
}

// Allocate a list of variables.
// res a;
// res a, b, c, ...;
func genir_reservation(r *Routine, node *parser.Node) error {
	for _, child := range node.Children {
		if child.Tok.Rule == parser.Rulevariable {
			if err := reserve_variable(r, child); err != nil {
				return err
			}
		}
	}
	return nil
}
func genir_returning(r *Routine, node *parser.Node) error {
	panic("Not Implemented!")
	return nil
}
func genir_assignment(r *Routine, node *parser.Node) error {
	panic("Not Implemented!")
	return nil
}
func genir_labeling(r *Routine, node *parser.Node) error {
	panic("Not Implemented!")
	return nil
}
func genir_jumping(r *Routine, node *parser.Node) error {
	panic("Not Implemented!")
	return nil
}

func genir_logicblock(r *Routine, node *parser.Node) error {
	//TODO(hunter): implement logic blocks
	panic("Not Implemented!")
	return nil
}
