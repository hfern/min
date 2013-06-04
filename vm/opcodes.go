package vm

type Opcode interface {
	Byte() byte
}

type Operation byte
type Register byte

const NUM_REGS int = 17

const (
	REGA Register = 0
	REGB Register = 1
	REGC Register = 2
	REGD Register = 3
	REGE Register = 4
	REGF          = 5
	REGG          = 6
	REGH          = 7
	REGI          = 8
	REGJ          = 9
	REGK          = 10
	REGL          = 11
	REGM          = 12
	REGN          = 13
	REGO          = 14
	REGP          = 15
	REGQ          = 16
)

const (
	OP                Operation = 0
	ERRNONE           Operation = 0
	ERRDONE           Operation = 1
	ERROPCODENOTFOUND Operation = 2
	ERRBREAK          Operation = 3
	BREAK_OUTOFBOUNDS Operation = 1
	END               Operation = 1
	SET               Operation = 2
	ADD               Operation = 3
	SUB               Operation = 4
	MUL               Operation = 5
	DIV               Operation = 6
	MOV               Operation = 7
	ADDREG            Operation = 8
	SUBREG            Operation = 9
	MULREG            Operation = 10
	DIVREG            Operation = 11
	MOD               Operation = 12
	JE                Operation = 13
	JNE               Operation = 14
	JL                Operation = 15
	JG                Operation = 16
	RELJE             Operation = 17
	RELJNE            Operation = 18
	RELJL             Operation = 19
	RELJG             Operation = 20
	STRPS             Operation = 21
	STPS              Operation = 22
	STPP              Operation = 23
	STPR              Operation = 24
	SQRT              Operation = 25
	NEG               Operation = 26
	SETL              Operation = 253
	NONE              Operation = 254
	BREAK             Operation = 255
)

func (op Operation) Byte() byte { return byte(op) }
func (reg Register) Byte() byte { return byte(reg) }
