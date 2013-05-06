package vm

type Opcode byte

const (
	OP Opcode = 0

	ERRNONE           = 0
	ERRDONE           = 1
	ERROPCODENOTFOUND = 2
	ERRBREAK          = 3

	BREAK_OUTOFBOUNDS = 1
	NUM_REGS          = 17

	REGA = 0
	REGB = 1
	REGC = 2
	REGD = 3
	REGE = 4
	REGF = 5
	REGG = 6
	REGH = 7
	REGI = 8
	REGJ = 9
	REGK = 10
	REGL = 11
	REGM = 12
	REGN = 13
	REGO = 14
	REGP = 15
	REGQ = 16

	END    = 1
	SET    = 2
	ADD    = 3
	SUB    = 4
	MUL    = 5
	DIV    = 6
	MOV    = 7
	ADDREG = 8
	SUBREG = 9
	MULREG = 10
	DIVREG = 11
	MOD    = 12

	JE  = 13
	JNE = 14
	JL  = 15
	JG  = 16

	RELJE  = 17
	RELJNE = 18
	RELJL  = 19
	RELJG  = 20

	STRPS = 21
	STPS  = 22
	STPP  = 23
	STPR  = 24

	SQRT = 25
	NEG  = 26

	SETL  = 253
	NONE  = 254
	BREAK = 255
)
