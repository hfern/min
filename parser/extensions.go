package parser

type Token interface {
	Begin() int
	End() int
	Next() int
	GetRule() Rule
}

func (t *token16) Begin() int    { return int(t.begin) }
func (t *token16) End() int      { return int(t.end) }
func (t *token16) Next() int     { return int(t.next) }
func (t *token16) GetRule() Rule { return t.Rule }

func (t *token32) Begin() int    { return int(t.begin) }
func (t *token32) End() int      { return int(t.end) }
func (t *token32) Next() int     { return int(t.next) }
func (t *token32) GetRule() Rule { return t.Rule }
