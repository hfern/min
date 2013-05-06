package compiler

type Routine struct {
	registers RegisterMap
	_program  *Program
	vmap      VariablePool
}

func NewRoutine() Routine {
	rout := Routine{}
	rout.registers = NewRegisterMap()
	rout.vmap = NewVariablePool()

	return rout
}
