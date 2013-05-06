package compiler

import (
	"github.com/hfern/min/vm"
)

type Register struct {
	id       uint
	reserved bool
}

func (r *Register) Lock() {
	r.reserved = true
}

func (r *Register) Locked() bool {
	return r.reserved
}

func (r *Register) Unlock() {
	r.reserved = false
}

type RegisterMap struct {
	registers [vm.NUM_REGS]*Register
}

func NewRegisterMap() RegisterMap {
	regs := RegisterMap{}
	for i, _ := range regs.registers {
		iu := uint(i)
		register := &Register{}
		register.id = iu
		register.Unlock()
		regs.registers[i] = register
	}
	return regs
}

func (c *RegisterMap) ReserveRegister() *Register {
	for _, reg := range c.registers {
		if !reg.Locked() {
			reg.Lock()
			return reg
		}
	}
	return nil
}

func (c *RegisterMap) RegistersInUse() []*Register {
	inuse := make([]*Register, 0, 5)
	for _, reg := range c.registers {
		if !reg.Locked() {
			inuse = append(inuse, reg)
		}
	}
	return inuse
}
