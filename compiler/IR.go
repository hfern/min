package compiler

import (
	"github.com/hfern/min/parser"
	"github.com/hfern/min/vm"
)

type IRSegment interface {
	Size() int // Size of Segments in bytes
	Pass(IRContext) bool
	Emit() []byte
}

type IRArray []IRSegment

type IRContext struct {
	PassNumber int
	Program    *Program
	IRArray    *IRArray
}

type IRLiteral struct {
	code []byte
}

type IRLabel struct {
	symbol string
}

// Represents the 4byte address of the jump,
// not the jump code itself
//    MOVL B
// -> 0 0 0 0 
//    JE B, B, B
type IRJump struct {
	symbol    string
	_position int // jumplocation
}

type IRFuncCall struct {
	returnreg vm.Register
	args      []vm.Register
	target    *Routine
	caller    *Routine
	calltok   *parser.Node
	__jumpto  int
}

func NewIRArray() IRArray { return make([]IRSegment, 0) }
func (ir *IRArray) Add(segments ...IRSegment) {
	for _, seg := range segments {
		*ir = append(*ir, seg)
	}
}

func (l *IRLiteral) Size() int {
	return len(l.code)
}
func (l *IRLiteral) Pass(_ IRContext) bool {
	return true
}
func (l *IRLiteral) Emit() []byte {
	return l.code
}

func (l *IRLabel) Size() int {
	return 0
}
func (l *IRLabel) Pass(_ IRContext) bool {
	return true
}
func (l *IRLabel) Emit() []byte {
	return []byte{}
}

func (j *IRJump) Size() int {
	return 4
}
func (j *IRJump) Pass(ctx IRContext) bool {
	if ctx.PassNumber == 1 {
		// TODO: Make this work.
		// function should set _position to byte value of
		// respective symbol
		return true
	}
	return false
}
func (j *IRJump) Emit() []byte {
	buf := make([]byte, 0, 4)
	setL(&buf, j._position)
	return buf
}

func (l *IRFuncCall) Size() int {
	// Push
	variables_to_pack := l.target.vmap.VariablesAlive(l.calltok.Tok.End())
	pack := (1 + 1) * len(variables_to_pack) // eachreg: STPR %reg

	push_return_location := 1 + 1     // VM_OPSTRPS $returnreg
	push_return_location += 1 + 1 + 1 // ADD $returnreg 14 // jump forward additional 14 bytes
	push_return_location += 1 + 1     // STPR $returnreg
	jump := 1 + 1 + 4                 // SETL $returnreg %jumploc{4 bytes}
	jump += 1 + 1 + 1                 // JE $returnreg $returnreg $returnreg

	unpack := (1 + 1) * len(variables_to_pack) // eachreg: STPP %reg
	return pack + push_return_location + jump + unpack
}
func (l *IRFuncCall) Pass(ctx IRContext) bool {
	if ctx.PassNumber == 2 {
		l.__jumpto = 0
		for _, ir := range ([]IRSegment)(*ctx.IRArray) {
			switch raw := (ir).(type) {
			case *IRLabel:
				if raw.symbol == l.target.Symbol() {
					return true
				}
			default:
				l.__jumpto += ir.Size()
				break
			}
		}
	}
	if ctx.PassNumber >= 2 {
		return true
	}
	return false
}
func (l *IRFuncCall) Emit() []byte {
	codesegment := make([]byte, 0, l.Size())

	variables_to_pack := l.target.vmap.VariablesAlive(l.calltok.Tok.End())
	for _, variable := range variables_to_pack {
		byteadd(&codesegment, vm.STPR, byte(variable.register.id))
	}

	retreg := l.returnreg

	byteadd(&codesegment, vm.STRPS, retreg)
	byteadd(&codesegment, vm.ADD, retreg, byte(14))
	byteadd(&codesegment, vm.STPR, retreg)
	byteadd(&codesegment, vm.SETL, retreg)
	setL(&codesegment, l.__jumpto)
	byteadd(&codesegment, vm.JE, retreg, retreg, retreg)

	for i := len(variables_to_pack) - 1; i <= 0; i-- {
		byteadd(&codesegment, vm.STPP, variables_to_pack[i].register.id)
	}
	return codesegment
}
