package compiler

import (
	"errors"
	"fmt"
	"github.com/hfern/min/parser"
	"strings"
)

func err_routine_already_exists(p *Program, oldr, newr *Routine) error {
	errstr := fmt.Sprintf(
		"Error at line %d: routine \"%s\" already defined at line %d.",
		line_no(&p.sourcecode, newr.__node.Tok.Begin()),
		newr.GetName(),
		line_no(&p.sourcecode, oldr.__node.Tok.Begin()),
	)
	return errors.New(errstr)
}

func newError(reasons ...interface{}) error {
	return errors.New(fmt.Sprint(reasons...))
}

func errorExpectingOneOf(tok parser.State16, src *string, expected []parser.Rule) error {
	expected_str := make([]string, 0, len(expected))
	for i, expect := range expected {
		expected_str[i] = parser.Rul3s[expect]
	}
	return newError(
		"Unexpected ",
		parser.Rul3s[tok.Rule],
		" at line ",
		line_no(src, tok.Begin()),
		". Expected one of (",
		strings.Join(expected_str, ","),
		").",
	)
}

func errorCannotReserveRegister(r *Routine, variable string, node *parser.Node) error {
	return newError(
		"Cannot reserve an unused register for use by variable \"",
		variable,
		"\" at line ",
		line_no(&r.__program.sourcecode, node.Tok.Begin()),
		"! (Try splitting program into more functions.)",
	)
}
