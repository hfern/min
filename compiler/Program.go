package compiler

type Program struct {
	symbols    SymbolMap
	__compiler *Compiler
	routines   map[string]*Routine
}

func NewProgram() Program {
	p := Program{}
	p.symbols = NewSymbolMap()
	p.routines = make(map[string]*Routine)
	return p
}
