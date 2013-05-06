package parser

import ()

type tokenpool struct {
	pool  []State16
	index int
}

func (p *tokenpool) Enumerate(raw <-chan State16, filterwhitespace bool) {
	p.index = -1
	p.pool = make([]State16, 0, 64)

	var use <-chan State16
	if filterwhitespace {
		use = whitespacefilter(raw)
	} else {
		use = raw
	}

	for tok := range use {
		p.pool = append(p.pool, tok)
	}
}

func (p *tokenpool) _get(n int) *State16 {
	if n < len(p.pool) {
		logPoolAccess(p, n, true)
		return &p.pool[n]
	}
	logPoolAccess(p, n, false)
	return nil
}

func (p *tokenpool) Peek() *State16 {
	return p._get(p.index + 1)
}

func (p *tokenpool) Next() *State16 {
	p.index++
	return p._get(p.index)
}

func (p *tokenpool) Back() {
	p.index--
}
