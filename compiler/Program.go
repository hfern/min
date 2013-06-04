package compiler

type Program struct {
	symbols         SymbolMap
	__compiler      *Compiler
	routines        []*Routine
	routinesByNames map[string]*Routine
	sourcecode      string
}

func NewProgram() Program {
	p := Program{}
	p.symbols = NewSymbolMap()
	p.routines = make([]*Routine, 0)
	p.routinesByNames = make(map[string]*Routine)
	return p
}

func (p *Program) RoutineExists(name string) bool {
	if _, ok := p.routinesByNames[name]; ok {
		return true
	}
	return false
}

func (p *Program) addRoutine(rout *Routine) {
	p.routines = append(p.routines, rout)
}

func (p *Program) linkRoutine(rout *Routine) error {
	name := rout.GetName()
	if p.RoutineExists(name) {
		return err_routine_already_exists(p, p.routinesByNames[name], rout)
	}
	p.routinesByNames[name] = rout
	return nil
}
