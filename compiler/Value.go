package compiler

type ValueType byte

const (
	ValueConstant ValueType = iota
	ValueVariable
)

type Value struct {
	Type     ValueType
	Constant int
	Register *Register
}
