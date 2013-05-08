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

func (t *State16) Begin() int    { return t.token16.Begin() }
func (t *State16) End() int      { return t.token16.End() }
func (t *State16) Next() int     { return t.token16.Next() }
func (t *State16) GetRule() Rule { return t.token16.GetRule() }

func (t *State32) Begin() int    { return t.token32.Begin() }
func (t *State32) End() int      { return t.token32.End() }
func (t *State32) Next() int     { return t.token32.Next() }
func (t *State32) GetRule() Rule { return t.token32.GetRule() }
