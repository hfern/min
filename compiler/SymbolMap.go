package compiler

type SymbolMap struct {
	_map map[string]interface{}
}

func NewSymbolMap() SymbolMap {
	syms := SymbolMap{}
	syms._map = make(map[string]interface{}, 10)
	return syms
}

func (m *SymbolMap) Add(sym string, val interface{}) {
	m._map[sym] = val
}

func (m *SymbolMap) Exists(sym string) bool {
	if _, ok := m._map[sym]; ok {
		return true
	}
	return false
}
