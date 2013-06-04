package compiler

import (
	"bytes"
	"encoding/binary"
	"github.com/hfern/min/parser"
	"strings"
)

type ruleHandler struct {
	rule    parser.Rule
	handler func(*Routine, *parser.Node) error
}

func line_no(text *string, position int) int {
	if len(*text) <= position {
		return -1
	}
	return strings.Count((*text)[0:position], "\n") + 1
}

// setL appends a 4 byte integer to the by byte array in 
// network byte order
func setL(bytearray *[]byte, number int) {
	buf := bytes.NewBuffer(make([]byte, 0, 4))
	binary.Write(buf, binary.BigEndian, uint32(number))
	for _, b := range buf.Bytes() {
		*bytearray = append(*bytearray, b)
	}
}

type byteable interface {
	Byte() byte
}

func byteadd(bytearray *[]byte, additionalbytes ...interface{}) {
	for _, by := range additionalbytes {
		switch typ := by.(type) {
		case byte:
			*bytearray = append(*bytearray, typ)
			break
		default:
			*bytearray = append(*bytearray, typ.(byteable).Byte())
			break
		}
	}
}

func handlepairs(rout *Routine, node *parser.Node, pairs []ruleHandler) error {
	for _, pair := range pairs {
		if node.Tok.Rule == pair.rule {
			return pair.handler(rout, node)
		}
	}

	expected := make([]parser.Rule, 0, len(pairs))
	for i, pair := range pairs {
		expected[i] = pair.rule
	}
	return errorExpectingOneOf(node.Tok, &rout.__program.sourcecode, expected)
}
