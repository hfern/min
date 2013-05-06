package compiler

import (
	"github.com/hfern/min/parser"
)

/**
 * VariablePool manages various variable-related tasks.
 * For instance, it keeps track of variable positions within the 
 * token tree to detect when it can free registers.
 */

type VariableMeta struct {
	name      string
	locations []*parser.State16
	register  *Register
	allocated bool
}

func NewVariableMeta() VariableMeta {
	v := VariableMeta{allocated: false}
	v.locations = make([]*parser.State16, 0, 5)
	return v
}

func (v *VariableMeta) Allocate() {
	//@TODO Have this write to the opcode stream the allocation instructuctions
	v.allocated = true
}

func (v *VariableMeta) Deallocate() {
	//@TODO see VariableMeta.Allocate
	v.register.Unlock()
	v.allocated = false
}

func (v *VariableMeta) Allocated() bool {
	return v.allocated
}

type VariablePool struct {
	_map map[string]*VariableMeta
}

func NewVariablePool() VariablePool {
	p := VariablePool{}
	p._map = make(map[string]*VariableMeta, 5)
	return p
}

func (pool *VariablePool) Exists(name string) bool {
	if _, ok := pool._map[name]; ok {
		return true
	}
	return false
}

func (p *VariablePool) AddInstance(v string, tok *parser.State16, regs *RegisterMap) {
	if !p.Exists(v) {
		nw := NewVariableMeta()
		nw.name = v
		nw.register = regs.ReserveRegister()
		p._map[v] = &nw
	}
	p._map[v].locations = append(p._map[v].locations, tok)
}

/**
 * Free registers of variables no longer referenced.
 */
func (p *VariablePool) Free(location parser.State16) {
	for _, v := range p._map {
		index := len(v.locations) - 1
		if index < 0 {
			continue
		}
		lastloc := v.locations[index].End()
		if lastloc < location.End() {
			// Last reference in routine was farther 
			// than current scan location
			v.Deallocate()
		}
	}
}
