package compiler

import (
	"flag"
	"github.com/hfern/min/parser"
	"log"
)

var flag_vvv *bool = flag.Bool("cmp-vvv", false, "Very, Very Verbose compiler logging.")

func log_function_call_saved(routine_name string, node *parser.Node) {
	if !*flag_vvv {
		return
	}
	log.Printf("Excluded %s:%s from variable positions (is function call).", routine_name, node.Source())
}

func log_number_funccalls_saved(routine_name string, number int) {
	if !*flag_vvv {
		return
	}
	log.Printf("Excluded %d function calls from routine:%s's variable positions tracing.", number, routine_name)
}

func log_variable_trace(routine_name string, node *parser.Node) {
	if !*flag_vvv {
		return
	}
	log.Printf("VTrace:%s: %s @ %d", routine_name, node.Source(), node.Tok.Begin())
}
