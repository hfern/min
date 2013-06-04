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
	locations []parser.State16
	register  *Register
	allocated bool
}

func NewVariableMeta() VariableMeta {
	v := VariableMeta{allocated: false}
	v.locations = make([]parser.State16, 0, 5)
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

func (p *VariablePool) AddInstance(varnode *parser.Node) {
	strvalue := varnode.Source()
	if !p.Exists(strvalue) {
		nw := NewVariableMeta()
		nw.name = strvalue
		p._map[strvalue] = &nw
	}
	p._map[strvalue].locations = append(p._map[strvalue].locations, varnode.Tok)
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

// VariablesAlive returns all variables that have been allocated
// before position and whose values will be used after position.
// Mathematically, is VarsBefore Intersect VarsAfter
// TODO(hunter): This algo has a terrible time complexity of n!
func (p *VariablePool) VariablesAlive(position int) []*VariableMeta {
	before := p.VariablesBefore(position)
	after := p.VariablesAfter(position)
	alive := make([]*VariableMeta, 0, 5)

	for _, var1 := range before {
		for _, var2 := range after {
			if var1 == var2 {
				alive = append(alive, var1)
				break
			}
		}
	}

	return alive
}

func (p *VariablePool) VariablesBefore(position int) []*VariableMeta {
	existingvars := make([]*VariableMeta, 0, 5)
	for _, variable := range p._map {
		for _, state := range variable.locations {
			if state.End() < position {
				existingvars = append(existingvars, variable)
				break
			}
		}
	}
	return existingvars
}

func (p *VariablePool) VariablesAfter(position int) []*VariableMeta {
	existingvars := make([]*VariableMeta, 0, 5)
	for _, variable := range p._map {
		for _, state := range variable.locations {
			if state.End() > position {
				existingvars = append(existingvars, variable)
				break
			}
		}
	}
	return existingvars
}
