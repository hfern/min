package parser

import (
	/*"bytes"*/
	"fmt"
	"math"
	"sort"
	"strconv"
)

const END_SYMBOL byte = 0

/* The rule types inferred from the grammar are below. */
type Rule uint8

const (
	RuleUnknown Rule = iota
	Ruleprogram
	Ruleroutine
	Ruleoperation
	Ruleopaction
	Rulereservation
	Rulereturning
	Ruleassignment
	Rulelabeling
	Rulejumping
	Rulevalue
	Ruleexpr
	Rulefunccall
	Rulecodestatement
	Rulecodeblock
	Rulelogicblock
	Ruleifblock
	Ruleelseblock
	Rulevariable
	Ruleparamaterdecl
	Rulecallparams
	Ruleparameters
	Rulecomma
	Rulekwreserve
	Rulekwreturn
	Rulekwroutine
	Rulekwjump
	Rulekwlabel
	Rulekwif
	Rulekwelse
	Ruletokadd
	Ruletoksub
	Ruletokmul
	Ruletokdiv
	Ruleendl
	Rulerawmath
	Rulemath
	Rulepopen
	Rulepclose
	Ruletoklt
	Ruletokgt
	Ruletokeq
	Ruletokle
	Ruletokge
	Ruletokne
	Rulecomparisontoken
	Rulecomparison
	Rulecomparison_paren
	Rulepositivenum
	Rulenegativenum
	Rulenumber
	Ruledigit
	Rulecommentblock
	Rulecommentdoubleslash
	Rulecomment
	Ruleliteralspace
	Rulespace
	Ruleminspace
	Ruleoptspace

	RulePre_
	Rule_In_
	Rule_Suf
)

var Rul3s = [...]string{
	"Unknown",
	"program",
	"routine",
	"operation",
	"opaction",
	"reservation",
	"returning",
	"assignment",
	"labeling",
	"jumping",
	"value",
	"expr",
	"funccall",
	"codestatement",
	"codeblock",
	"logicblock",
	"ifblock",
	"elseblock",
	"variable",
	"paramaterdecl",
	"callparams",
	"parameters",
	"comma",
	"kwreserve",
	"kwreturn",
	"kwroutine",
	"kwjump",
	"kwlabel",
	"kwif",
	"kwelse",
	"tokadd",
	"toksub",
	"tokmul",
	"tokdiv",
	"endl",
	"rawmath",
	"math",
	"popen",
	"pclose",
	"toklt",
	"tokgt",
	"tokeq",
	"tokle",
	"tokge",
	"tokne",
	"comparisontoken",
	"comparison",
	"comparison_paren",
	"positivenum",
	"negativenum",
	"number",
	"digit",
	"commentblock",
	"commentdoubleslash",
	"comment",
	"literalspace",
	"space",
	"minspace",
	"optspace",

	"Pre_",
	"_In_",
	"_Suf",
}

type TokenTree interface {
	Print()
	PrintSyntax()
	PrintSyntaxTree(buffer string)
	Add(rule Rule, begin, end, next, depth int)
	Expand(index int) TokenTree
	Tokens() <-chan token32
	Error() []token32
	trim(length int)
}

/* ${@} bit structure for abstract syntax tree */
type token16 struct {
	Rule
	begin, end, next int16
}

func (t *token16) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token16) isParentOf(u token16) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token16) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token16) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
}

type tokens16 struct {
	tree    []token16
	ordered [][]token16
}

func (t *tokens16) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens16) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens16) Order() [][]token16 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int16, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.Rule == RuleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token16, len(depths)), make([]token16, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int16(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State16 struct {
	token16
	depths []int16
	leaf   bool
}

func (t *tokens16) PreOrder() (<-chan State16, [][]token16) {
	s, ordered := make(chan State16, 6), t.Order()
	go func() {
		var states [8]State16
		for i, _ := range states {
			states[i].depths = make([]int16, len(ordered))
		}
		depths, state, depth := make([]int16, len(ordered)), 0, 1
		write := func(t token16, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int16(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token16 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token16{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token16{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token16{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens16) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens16) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens16) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token16{Rule: rule, begin: int16(begin), end: int16(end), next: int16(depth)}
}

func (t *tokens16) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens16) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

/* ${@} bit structure for abstract syntax tree */
type token32 struct {
	Rule
	begin, end, next int32
}

func (t *token32) isZero() bool {
	return t.Rule == RuleUnknown && t.begin == 0 && t.end == 0 && t.next == 0
}

func (t *token32) isParentOf(u token32) bool {
	return t.begin <= u.begin && t.end >= u.end && t.next > u.next
}

func (t *token32) GetToken32() token32 {
	return token32{Rule: t.Rule, begin: int32(t.begin), end: int32(t.end), next: int32(t.next)}
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v %v", Rul3s[t.Rule], t.begin, t.end, t.next)
}

type tokens32 struct {
	tree    []token32
	ordered [][]token32
}

func (t *tokens32) trim(length int) {
	t.tree = t.tree[0:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) Order() [][]token32 {
	if t.ordered != nil {
		return t.ordered
	}

	depths := make([]int32, 1, math.MaxInt16)
	for i, token := range t.tree {
		if token.Rule == RuleUnknown {
			t.tree = t.tree[:i]
			break
		}
		depth := int(token.next)
		if length := len(depths); depth >= length {
			depths = depths[:depth+1]
		}
		depths[depth]++
	}
	depths = append(depths, 0)

	ordered, pool := make([][]token32, len(depths)), make([]token32, len(t.tree)+len(depths))
	for i, depth := range depths {
		depth++
		ordered[i], pool, depths[i] = pool[:depth], pool[depth:], 0
	}

	for i, token := range t.tree {
		depth := token.next
		token.next = int32(i)
		ordered[depth][depths[depth]] = token
		depths[depth]++
	}
	t.ordered = ordered
	return ordered
}

type State32 struct {
	token32
	depths []int32
	leaf   bool
}

func (t *tokens32) PreOrder() (<-chan State32, [][]token32) {
	s, ordered := make(chan State32, 6), t.Order()
	go func() {
		var states [8]State32
		for i, _ := range states {
			states[i].depths = make([]int32, len(ordered))
		}
		depths, state, depth := make([]int32, len(ordered)), 0, 1
		write := func(t token32, leaf bool) {
			S := states[state]
			state, S.Rule, S.begin, S.end, S.next, S.leaf = (state+1)%8, t.Rule, t.begin, t.end, int32(depth), leaf
			copy(S.depths, depths)
			s <- S
		}

		states[state].token32 = ordered[0][0]
		depths[0]++
		state++
		a, b := ordered[depth-1][depths[depth-1]-1], ordered[depth][depths[depth]]
	depthFirstSearch:
		for {
			for {
				if i := depths[depth]; i > 0 {
					if c, j := ordered[depth][i-1], depths[depth-1]; a.isParentOf(c) &&
						(j < 2 || !ordered[depth-1][j-2].isParentOf(c)) {
						if c.end != b.begin {
							write(token32{Rule: Rule_In_, begin: c.end, end: b.begin}, true)
						}
						break
					}
				}

				if a.begin < b.begin {
					write(token32{Rule: RulePre_, begin: a.begin, end: b.begin}, true)
				}
				break
			}

			next := depth + 1
			if c := ordered[next][depths[next]]; c.Rule != RuleUnknown && b.isParentOf(c) {
				write(b, false)
				depths[depth]++
				depth, a, b = next, b, c
				continue
			}

			write(b, true)
			depths[depth]++
			c, parent := ordered[depth][depths[depth]], true
			for {
				if c.Rule != RuleUnknown && a.isParentOf(c) {
					b = c
					continue depthFirstSearch
				} else if parent && b.end != a.end {
					write(token32{Rule: Rule_Suf, begin: b.end, end: a.end}, true)
				}

				depth--
				if depth > 0 {
					a, b, c = ordered[depth-1][depths[depth-1]-1], a, ordered[depth][depths[depth]]
					parent = a.isParentOf(b)
					continue
				}

				break depthFirstSearch
			}
		}

		close(s)
	}()
	return s, ordered
}

func (t *tokens32) PrintSyntax() {
	tokens, ordered := t.PreOrder()
	max := -1
	for token := range tokens {
		if !token.leaf {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[36m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[36m%v\x1B[m\n", Rul3s[token.Rule])
		} else if token.begin == token.end {
			fmt.Printf("%v", token.begin)
			for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
				fmt.Printf(" \x1B[31m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
			}
			fmt.Printf(" \x1B[31m%v\x1B[m\n", Rul3s[token.Rule])
		} else {
			for c, end := token.begin, token.end; c < end; c++ {
				if i := int(c); max+1 < i {
					for j := max; j < i; j++ {
						fmt.Printf("skip %v %v\n", j, token.String())
					}
					max = i
				} else if i := int(c); i <= max {
					for j := i; j <= max; j++ {
						fmt.Printf("dupe %v %v\n", j, token.String())
					}
				} else {
					max = int(c)
				}
				fmt.Printf("%v", c)
				for i, leaf, depths := 0, int(token.next), token.depths; i < leaf; i++ {
					fmt.Printf(" \x1B[34m%v\x1B[m", Rul3s[ordered[i][depths[i]-1].Rule])
				}
				fmt.Printf(" \x1B[34m%v\x1B[m\n", Rul3s[token.Rule])
			}
			fmt.Printf("\n")
		}
	}
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	tokens, _ := t.PreOrder()
	for token := range tokens {
		for c := 0; c < int(token.next); c++ {
			fmt.Printf(" ")
		}
		fmt.Printf("\x1B[34m%v\x1B[m %v\n", Rul3s[token.Rule], strconv.Quote(buffer[token.begin:token.end]))
	}
}

func (t *tokens32) Add(rule Rule, begin, end, depth, index int) {
	t.tree[index] = token32{Rule: rule, begin: int32(begin), end: int32(end), next: int32(depth)}
}

func (t *tokens32) Tokens() <-chan token32 {
	s := make(chan token32, 16)
	go func() {
		for _, v := range t.tree {
			s <- v.GetToken32()
		}
		close(s)
	}()
	return s
}

func (t *tokens32) Error() []token32 {
	ordered := t.Order()
	length := len(ordered)
	tokens, length := make([]token32, length), length-1
	for i, _ := range tokens {
		o := ordered[length-i]
		if len(o) > 1 {
			tokens[i] = o[len(o)-2].GetToken32()
		}
	}
	return tokens
}

func (t *tokens16) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		for i, v := range tree {
			expanded[i] = v.GetToken32()
		}
		return &tokens32{tree: expanded}
	}
	return nil
}

func (t *tokens32) Expand(index int) TokenTree {
	tree := t.tree
	if index >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	return nil
}

type VMTree struct {

	//Expression

	Buffer string
	rules  [59]func() bool
	Parse  func(rule ...int) error
	Reset  func()
	TokenTree
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer string, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer[0:] {
		if c == '\n' {
			line, symbol = line+1, 0
		} else {
			symbol++
		}
		if i == positions[j] {
			translations[positions[j]] = textPosition{line, symbol}
			for j++; j < length; j++ {
				if i != positions[j] {
					continue search
				}
			}
			break search
		}
	}

	return translations
}

type parseError struct {
	p *VMTree
}

func (e *parseError) Error() string {
	tokens, error := e.p.TokenTree.Error(), "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.Buffer, positions)
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		error += fmt.Sprintf("parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n",
			Rul3s[token.Rule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			/*strconv.Quote(*/ e.p.Buffer[begin:end] /*)*/)
	}

	return error
}

func (p *VMTree) PrintSyntaxTree() {
	p.TokenTree.PrintSyntaxTree(p.Buffer)
}

func (p *VMTree) Highlighter() {
	p.TokenTree.PrintSyntax()
}

func (p *VMTree) Init() {
	if p.Buffer[len(p.Buffer)-1] != END_SYMBOL {
		p.Buffer = p.Buffer + string(END_SYMBOL)
	}

	var tree TokenTree = &tokens16{tree: make([]token16, math.MaxInt16)}
	position, depth, tokenIndex, buffer, rules := 0, 0, 0, p.Buffer, p.rules

	p.Parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.TokenTree = tree
		if matches {
			p.TokenTree.trim(tokenIndex)
			return nil
		}
		return &parseError{p}
	}

	p.Reset = func() {
		position, tokenIndex, depth = 0, 0, 0
	}

	add := func(rule Rule, begin int) {
		if t := tree.Expand(tokenIndex); t != nil {
			tree = t
		}
		tree.Add(rule, begin, position, depth, tokenIndex)
		tokenIndex++
	}

	matchDot := func() bool {
		if buffer[position] != END_SYMBOL {
			position++
			return true
		}
		return false
	}

	/*matchChar := func(c byte) bool {
		if buffer[position] == c {
			position++
			return true
		}
		return false
	}*/

	/*matchRange := func(lower byte, upper byte) bool {
		if c := buffer[position]; c >= lower && c <= upper {
			position++
			return true
		}
		return false
	}*/

	rules = [...]func() bool{
		nil,
		/* 0 program <- <(optspace (routine optspace)+ optspace)> */
		func() bool {
			position0, tokenIndex0, depth0 := position, tokenIndex, depth
			{
				position1 := position
				depth++
				if !rules[Ruleoptspace]() {
					goto l0
				}
				if !rules[Ruleroutine]() {
					goto l0
				}
				if !rules[Ruleoptspace]() {
					goto l0
				}
			l2:
				{
					position3, tokenIndex3, depth3 := position, tokenIndex, depth
					if !rules[Ruleroutine]() {
						goto l3
					}
					if !rules[Ruleoptspace]() {
						goto l3
					}
					goto l2
				l3:
					position, tokenIndex, depth = position3, tokenIndex3, depth3
				}
				if !rules[Ruleoptspace]() {
					goto l0
				}
				depth--
				add(Ruleprogram, position1)
			}
			return true
		l0:
			position, tokenIndex, depth = position0, tokenIndex0, depth0
			return false
		},
		/* 1 routine <- <(kwroutine minspace variable optspace paramaterdecl codeblock)> */
		func() bool {
			position4, tokenIndex4, depth4 := position, tokenIndex, depth
			{
				position5 := position
				depth++
				if !rules[Rulekwroutine]() {
					goto l4
				}
				if !rules[Ruleminspace]() {
					goto l4
				}
				if !rules[Rulevariable]() {
					goto l4
				}
				if !rules[Ruleoptspace]() {
					goto l4
				}
				if !rules[Ruleparamaterdecl]() {
					goto l4
				}
				if !rules[Rulecodeblock]() {
					goto l4
				}
				depth--
				add(Ruleroutine, position5)
			}
			return true
		l4:
			position, tokenIndex, depth = position4, tokenIndex4, depth4
			return false
		},
		/* 2 operation <- <(opaction optspace endl)> */
		func() bool {
			position6, tokenIndex6, depth6 := position, tokenIndex, depth
			{
				position7 := position
				depth++
				if !rules[Ruleopaction]() {
					goto l6
				}
				if !rules[Ruleoptspace]() {
					goto l6
				}
				if !rules[Ruleendl]() {
					goto l6
				}
				depth--
				add(Ruleoperation, position7)
			}
			return true
		l6:
			position, tokenIndex, depth = position6, tokenIndex6, depth6
			return false
		},
		/* 3 opaction <- <(reservation / returning / assignment / labeling / jumping)> */
		func() bool {
			position8, tokenIndex8, depth8 := position, tokenIndex, depth
			{
				position9 := position
				depth++
				{
					position10, tokenIndex10, depth10 := position, tokenIndex, depth
					if !rules[Rulereservation]() {
						goto l11
					}
					goto l10
				l11:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
					if !rules[Rulereturning]() {
						goto l12
					}
					goto l10
				l12:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
					if !rules[Ruleassignment]() {
						goto l13
					}
					goto l10
				l13:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
					if !rules[Rulelabeling]() {
						goto l14
					}
					goto l10
				l14:
					position, tokenIndex, depth = position10, tokenIndex10, depth10
					if !rules[Rulejumping]() {
						goto l8
					}
				}
			l10:
				depth--
				add(Ruleopaction, position9)
			}
			return true
		l8:
			position, tokenIndex, depth = position8, tokenIndex8, depth8
			return false
		},
		/* 4 reservation <- <(kwreserve minspace variable (comma variable)*)> */
		func() bool {
			position15, tokenIndex15, depth15 := position, tokenIndex, depth
			{
				position16 := position
				depth++
				if !rules[Rulekwreserve]() {
					goto l15
				}
				if !rules[Ruleminspace]() {
					goto l15
				}
				if !rules[Rulevariable]() {
					goto l15
				}
			l17:
				{
					position18, tokenIndex18, depth18 := position, tokenIndex, depth
					if !rules[Rulecomma]() {
						goto l18
					}
					if !rules[Rulevariable]() {
						goto l18
					}
					goto l17
				l18:
					position, tokenIndex, depth = position18, tokenIndex18, depth18
				}
				depth--
				add(Rulereservation, position16)
			}
			return true
		l15:
			position, tokenIndex, depth = position15, tokenIndex15, depth15
			return false
		},
		/* 5 returning <- <(kwreturn minspace expr)> */
		func() bool {
			position19, tokenIndex19, depth19 := position, tokenIndex, depth
			{
				position20 := position
				depth++
				if !rules[Rulekwreturn]() {
					goto l19
				}
				if !rules[Ruleminspace]() {
					goto l19
				}
				if !rules[Ruleexpr]() {
					goto l19
				}
				depth--
				add(Rulereturning, position20)
			}
			return true
		l19:
			position, tokenIndex, depth = position19, tokenIndex19, depth19
			return false
		},
		/* 6 assignment <- <(variable optspace '=' optspace expr optspace)> */
		func() bool {
			position21, tokenIndex21, depth21 := position, tokenIndex, depth
			{
				position22 := position
				depth++
				if !rules[Rulevariable]() {
					goto l21
				}
				if !rules[Ruleoptspace]() {
					goto l21
				}
				if buffer[position] != '=' {
					goto l21
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l21
				}
				if !rules[Ruleexpr]() {
					goto l21
				}
				if !rules[Ruleoptspace]() {
					goto l21
				}
				depth--
				add(Ruleassignment, position22)
			}
			return true
		l21:
			position, tokenIndex, depth = position21, tokenIndex21, depth21
			return false
		},
		/* 7 labeling <- <(kwlabel minspace variable optspace)> */
		func() bool {
			position23, tokenIndex23, depth23 := position, tokenIndex, depth
			{
				position24 := position
				depth++
				if !rules[Rulekwlabel]() {
					goto l23
				}
				if !rules[Ruleminspace]() {
					goto l23
				}
				if !rules[Rulevariable]() {
					goto l23
				}
				if !rules[Ruleoptspace]() {
					goto l23
				}
				depth--
				add(Rulelabeling, position24)
			}
			return true
		l23:
			position, tokenIndex, depth = position23, tokenIndex23, depth23
			return false
		},
		/* 8 jumping <- <(kwjump minspace variable optspace)> */
		func() bool {
			position25, tokenIndex25, depth25 := position, tokenIndex, depth
			{
				position26 := position
				depth++
				if !rules[Rulekwjump]() {
					goto l25
				}
				if !rules[Ruleminspace]() {
					goto l25
				}
				if !rules[Rulevariable]() {
					goto l25
				}
				if !rules[Ruleoptspace]() {
					goto l25
				}
				depth--
				add(Rulejumping, position26)
			}
			return true
		l25:
			position, tokenIndex, depth = position25, tokenIndex25, depth25
			return false
		},
		/* 9 value <- <(funccall / number / variable)> */
		func() bool {
			position27, tokenIndex27, depth27 := position, tokenIndex, depth
			{
				position28 := position
				depth++
				{
					position29, tokenIndex29, depth29 := position, tokenIndex, depth
					if !rules[Rulefunccall]() {
						goto l30
					}
					goto l29
				l30:
					position, tokenIndex, depth = position29, tokenIndex29, depth29
					if !rules[Rulenumber]() {
						goto l31
					}
					goto l29
				l31:
					position, tokenIndex, depth = position29, tokenIndex29, depth29
					if !rules[Rulevariable]() {
						goto l27
					}
				}
			l29:
				depth--
				add(Rulevalue, position28)
			}
			return true
		l27:
			position, tokenIndex, depth = position27, tokenIndex27, depth27
			return false
		},
		/* 10 expr <- <(value / math)> */
		func() bool {
			position32, tokenIndex32, depth32 := position, tokenIndex, depth
			{
				position33 := position
				depth++
				{
					position34, tokenIndex34, depth34 := position, tokenIndex, depth
					if !rules[Rulevalue]() {
						goto l35
					}
					goto l34
				l35:
					position, tokenIndex, depth = position34, tokenIndex34, depth34
					if !rules[Rulemath]() {
						goto l32
					}
				}
			l34:
				depth--
				add(Ruleexpr, position33)
			}
			return true
		l32:
			position, tokenIndex, depth = position32, tokenIndex32, depth32
			return false
		},
		/* 11 funccall <- <(variable '(' optspace callparams? ')')> */
		func() bool {
			position36, tokenIndex36, depth36 := position, tokenIndex, depth
			{
				position37 := position
				depth++
				if !rules[Rulevariable]() {
					goto l36
				}
				if buffer[position] != '(' {
					goto l36
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l36
				}
				{
					position38, tokenIndex38, depth38 := position, tokenIndex, depth
					if !rules[Rulecallparams]() {
						goto l38
					}
					goto l39
				l38:
					position, tokenIndex, depth = position38, tokenIndex38, depth38
				}
			l39:
				if buffer[position] != ')' {
					goto l36
				}
				position++
				depth--
				add(Rulefunccall, position37)
			}
			return true
		l36:
			position, tokenIndex, depth = position36, tokenIndex36, depth36
			return false
		},
		/* 12 codestatement <- <(logicblock / operation)> */
		func() bool {
			position40, tokenIndex40, depth40 := position, tokenIndex, depth
			{
				position41 := position
				depth++
				{
					position42, tokenIndex42, depth42 := position, tokenIndex, depth
					if !rules[Rulelogicblock]() {
						goto l43
					}
					goto l42
				l43:
					position, tokenIndex, depth = position42, tokenIndex42, depth42
					if !rules[Ruleoperation]() {
						goto l40
					}
				}
			l42:
				depth--
				add(Rulecodestatement, position41)
			}
			return true
		l40:
			position, tokenIndex, depth = position40, tokenIndex40, depth40
			return false
		},
		/* 13 codeblock <- <(optspace '{' (optspace codestatement)* optspace '}' optspace)> */
		func() bool {
			position44, tokenIndex44, depth44 := position, tokenIndex, depth
			{
				position45 := position
				depth++
				if !rules[Ruleoptspace]() {
					goto l44
				}
				if buffer[position] != '{' {
					goto l44
				}
				position++
			l46:
				{
					position47, tokenIndex47, depth47 := position, tokenIndex, depth
					if !rules[Ruleoptspace]() {
						goto l47
					}
					if !rules[Rulecodestatement]() {
						goto l47
					}
					goto l46
				l47:
					position, tokenIndex, depth = position47, tokenIndex47, depth47
				}
				if !rules[Ruleoptspace]() {
					goto l44
				}
				if buffer[position] != '}' {
					goto l44
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l44
				}
				depth--
				add(Rulecodeblock, position45)
			}
			return true
		l44:
			position, tokenIndex, depth = position44, tokenIndex44, depth44
			return false
		},
		/* 14 logicblock <- <(ifblock (optspace elseblock)?)> */
		func() bool {
			position48, tokenIndex48, depth48 := position, tokenIndex, depth
			{
				position49 := position
				depth++
				if !rules[Ruleifblock]() {
					goto l48
				}
				{
					position50, tokenIndex50, depth50 := position, tokenIndex, depth
					if !rules[Ruleoptspace]() {
						goto l50
					}
					if !rules[Ruleelseblock]() {
						goto l50
					}
					goto l51
				l50:
					position, tokenIndex, depth = position50, tokenIndex50, depth50
				}
			l51:
				depth--
				add(Rulelogicblock, position49)
			}
			return true
		l48:
			position, tokenIndex, depth = position48, tokenIndex48, depth48
			return false
		},
		/* 15 ifblock <- <(kwif minspace comparison_paren codeblock)> */
		func() bool {
			position52, tokenIndex52, depth52 := position, tokenIndex, depth
			{
				position53 := position
				depth++
				if !rules[Rulekwif]() {
					goto l52
				}
				if !rules[Ruleminspace]() {
					goto l52
				}
				if !rules[Rulecomparison_paren]() {
					goto l52
				}
				if !rules[Rulecodeblock]() {
					goto l52
				}
				depth--
				add(Ruleifblock, position53)
			}
			return true
		l52:
			position, tokenIndex, depth = position52, tokenIndex52, depth52
			return false
		},
		/* 16 elseblock <- <(kwelse optspace codeblock)> */
		func() bool {
			position54, tokenIndex54, depth54 := position, tokenIndex, depth
			{
				position55 := position
				depth++
				if !rules[Rulekwelse]() {
					goto l54
				}
				if !rules[Ruleoptspace]() {
					goto l54
				}
				if !rules[Rulecodeblock]() {
					goto l54
				}
				depth--
				add(Ruleelseblock, position55)
			}
			return true
		l54:
			position, tokenIndex, depth = position54, tokenIndex54, depth54
			return false
		},
		/* 17 variable <- <(([a-z] / [A-Z])+ ([a-z] / [A-Z] / [0-9])*)> */
		func() bool {
			position56, tokenIndex56, depth56 := position, tokenIndex, depth
			{
				position57 := position
				depth++
				{
					position60, tokenIndex60, depth60 := position, tokenIndex, depth
					if c := buffer[position]; c < 'a' || c > 'z' {
						goto l61
					}
					position++
					goto l60
				l61:
					position, tokenIndex, depth = position60, tokenIndex60, depth60
					if c := buffer[position]; c < 'A' || c > 'Z' {
						goto l56
					}
					position++
				}
			l60:
			l58:
				{
					position59, tokenIndex59, depth59 := position, tokenIndex, depth
					{
						position62, tokenIndex62, depth62 := position, tokenIndex, depth
						if c := buffer[position]; c < 'a' || c > 'z' {
							goto l63
						}
						position++
						goto l62
					l63:
						position, tokenIndex, depth = position62, tokenIndex62, depth62
						if c := buffer[position]; c < 'A' || c > 'Z' {
							goto l59
						}
						position++
					}
				l62:
					goto l58
				l59:
					position, tokenIndex, depth = position59, tokenIndex59, depth59
				}
			l64:
				{
					position65, tokenIndex65, depth65 := position, tokenIndex, depth
					{
						position66, tokenIndex66, depth66 := position, tokenIndex, depth
						if c := buffer[position]; c < 'a' || c > 'z' {
							goto l67
						}
						position++
						goto l66
					l67:
						position, tokenIndex, depth = position66, tokenIndex66, depth66
						if c := buffer[position]; c < 'A' || c > 'Z' {
							goto l68
						}
						position++
						goto l66
					l68:
						position, tokenIndex, depth = position66, tokenIndex66, depth66
						if c := buffer[position]; c < '0' || c > '9' {
							goto l65
						}
						position++
					}
				l66:
					goto l64
				l65:
					position, tokenIndex, depth = position65, tokenIndex65, depth65
				}
				depth--
				add(Rulevariable, position57)
			}
			return true
		l56:
			position, tokenIndex, depth = position56, tokenIndex56, depth56
			return false
		},
		/* 18 paramaterdecl <- <('<' optspace parameters? '>')> */
		func() bool {
			position69, tokenIndex69, depth69 := position, tokenIndex, depth
			{
				position70 := position
				depth++
				if buffer[position] != '<' {
					goto l69
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l69
				}
				{
					position71, tokenIndex71, depth71 := position, tokenIndex, depth
					if !rules[Ruleparameters]() {
						goto l71
					}
					goto l72
				l71:
					position, tokenIndex, depth = position71, tokenIndex71, depth71
				}
			l72:
				if buffer[position] != '>' {
					goto l69
				}
				position++
				depth--
				add(Ruleparamaterdecl, position70)
			}
			return true
		l69:
			position, tokenIndex, depth = position69, tokenIndex69, depth69
			return false
		},
		/* 19 callparams <- <(value (comma value)* optspace)> */
		func() bool {
			position73, tokenIndex73, depth73 := position, tokenIndex, depth
			{
				position74 := position
				depth++
				if !rules[Rulevalue]() {
					goto l73
				}
			l75:
				{
					position76, tokenIndex76, depth76 := position, tokenIndex, depth
					if !rules[Rulecomma]() {
						goto l76
					}
					if !rules[Rulevalue]() {
						goto l76
					}
					goto l75
				l76:
					position, tokenIndex, depth = position76, tokenIndex76, depth76
				}
				if !rules[Ruleoptspace]() {
					goto l73
				}
				depth--
				add(Rulecallparams, position74)
			}
			return true
		l73:
			position, tokenIndex, depth = position73, tokenIndex73, depth73
			return false
		},
		/* 20 parameters <- <(variable (comma variable)* optspace)> */
		func() bool {
			position77, tokenIndex77, depth77 := position, tokenIndex, depth
			{
				position78 := position
				depth++
				if !rules[Rulevariable]() {
					goto l77
				}
			l79:
				{
					position80, tokenIndex80, depth80 := position, tokenIndex, depth
					if !rules[Rulecomma]() {
						goto l80
					}
					if !rules[Rulevariable]() {
						goto l80
					}
					goto l79
				l80:
					position, tokenIndex, depth = position80, tokenIndex80, depth80
				}
				if !rules[Ruleoptspace]() {
					goto l77
				}
				depth--
				add(Ruleparameters, position78)
			}
			return true
		l77:
			position, tokenIndex, depth = position77, tokenIndex77, depth77
			return false
		},
		/* 21 comma <- <(optspace ',' optspace)> */
		func() bool {
			position81, tokenIndex81, depth81 := position, tokenIndex, depth
			{
				position82 := position
				depth++
				if !rules[Ruleoptspace]() {
					goto l81
				}
				if buffer[position] != ',' {
					goto l81
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l81
				}
				depth--
				add(Rulecomma, position82)
			}
			return true
		l81:
			position, tokenIndex, depth = position81, tokenIndex81, depth81
			return false
		},
		/* 22 kwreserve <- <('r' 'e' 's')> */
		func() bool {
			position83, tokenIndex83, depth83 := position, tokenIndex, depth
			{
				position84 := position
				depth++
				if buffer[position] != 'r' {
					goto l83
				}
				position++
				if buffer[position] != 'e' {
					goto l83
				}
				position++
				if buffer[position] != 's' {
					goto l83
				}
				position++
				depth--
				add(Rulekwreserve, position84)
			}
			return true
		l83:
			position, tokenIndex, depth = position83, tokenIndex83, depth83
			return false
		},
		/* 23 kwreturn <- <('r' 'e' 't' 'u' 'r' 'n')> */
		func() bool {
			position85, tokenIndex85, depth85 := position, tokenIndex, depth
			{
				position86 := position
				depth++
				if buffer[position] != 'r' {
					goto l85
				}
				position++
				if buffer[position] != 'e' {
					goto l85
				}
				position++
				if buffer[position] != 't' {
					goto l85
				}
				position++
				if buffer[position] != 'u' {
					goto l85
				}
				position++
				if buffer[position] != 'r' {
					goto l85
				}
				position++
				if buffer[position] != 'n' {
					goto l85
				}
				position++
				depth--
				add(Rulekwreturn, position86)
			}
			return true
		l85:
			position, tokenIndex, depth = position85, tokenIndex85, depth85
			return false
		},
		/* 24 kwroutine <- <('r' 'o' 'u' 't' 'i' 'n' 'e')> */
		func() bool {
			position87, tokenIndex87, depth87 := position, tokenIndex, depth
			{
				position88 := position
				depth++
				if buffer[position] != 'r' {
					goto l87
				}
				position++
				if buffer[position] != 'o' {
					goto l87
				}
				position++
				if buffer[position] != 'u' {
					goto l87
				}
				position++
				if buffer[position] != 't' {
					goto l87
				}
				position++
				if buffer[position] != 'i' {
					goto l87
				}
				position++
				if buffer[position] != 'n' {
					goto l87
				}
				position++
				if buffer[position] != 'e' {
					goto l87
				}
				position++
				depth--
				add(Rulekwroutine, position88)
			}
			return true
		l87:
			position, tokenIndex, depth = position87, tokenIndex87, depth87
			return false
		},
		/* 25 kwjump <- <('j' 'u' 'm' 'p')> */
		func() bool {
			position89, tokenIndex89, depth89 := position, tokenIndex, depth
			{
				position90 := position
				depth++
				if buffer[position] != 'j' {
					goto l89
				}
				position++
				if buffer[position] != 'u' {
					goto l89
				}
				position++
				if buffer[position] != 'm' {
					goto l89
				}
				position++
				if buffer[position] != 'p' {
					goto l89
				}
				position++
				depth--
				add(Rulekwjump, position90)
			}
			return true
		l89:
			position, tokenIndex, depth = position89, tokenIndex89, depth89
			return false
		},
		/* 26 kwlabel <- <('l' 'a' 'b' 'e' 'l')> */
		func() bool {
			position91, tokenIndex91, depth91 := position, tokenIndex, depth
			{
				position92 := position
				depth++
				if buffer[position] != 'l' {
					goto l91
				}
				position++
				if buffer[position] != 'a' {
					goto l91
				}
				position++
				if buffer[position] != 'b' {
					goto l91
				}
				position++
				if buffer[position] != 'e' {
					goto l91
				}
				position++
				if buffer[position] != 'l' {
					goto l91
				}
				position++
				depth--
				add(Rulekwlabel, position92)
			}
			return true
		l91:
			position, tokenIndex, depth = position91, tokenIndex91, depth91
			return false
		},
		/* 27 kwif <- <('i' 'f')> */
		func() bool {
			position93, tokenIndex93, depth93 := position, tokenIndex, depth
			{
				position94 := position
				depth++
				if buffer[position] != 'i' {
					goto l93
				}
				position++
				if buffer[position] != 'f' {
					goto l93
				}
				position++
				depth--
				add(Rulekwif, position94)
			}
			return true
		l93:
			position, tokenIndex, depth = position93, tokenIndex93, depth93
			return false
		},
		/* 28 kwelse <- <('e' 'l' 's' 'e')> */
		func() bool {
			position95, tokenIndex95, depth95 := position, tokenIndex, depth
			{
				position96 := position
				depth++
				if buffer[position] != 'e' {
					goto l95
				}
				position++
				if buffer[position] != 'l' {
					goto l95
				}
				position++
				if buffer[position] != 's' {
					goto l95
				}
				position++
				if buffer[position] != 'e' {
					goto l95
				}
				position++
				depth--
				add(Rulekwelse, position96)
			}
			return true
		l95:
			position, tokenIndex, depth = position95, tokenIndex95, depth95
			return false
		},
		/* 29 tokadd <- <('+' optspace)> */
		func() bool {
			position97, tokenIndex97, depth97 := position, tokenIndex, depth
			{
				position98 := position
				depth++
				if buffer[position] != '+' {
					goto l97
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l97
				}
				depth--
				add(Ruletokadd, position98)
			}
			return true
		l97:
			position, tokenIndex, depth = position97, tokenIndex97, depth97
			return false
		},
		/* 30 toksub <- <('-' optspace)> */
		func() bool {
			position99, tokenIndex99, depth99 := position, tokenIndex, depth
			{
				position100 := position
				depth++
				if buffer[position] != '-' {
					goto l99
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l99
				}
				depth--
				add(Ruletoksub, position100)
			}
			return true
		l99:
			position, tokenIndex, depth = position99, tokenIndex99, depth99
			return false
		},
		/* 31 tokmul <- <('*' space)> */
		func() bool {
			position101, tokenIndex101, depth101 := position, tokenIndex, depth
			{
				position102 := position
				depth++
				if buffer[position] != '*' {
					goto l101
				}
				position++
				if !rules[Rulespace]() {
					goto l101
				}
				depth--
				add(Ruletokmul, position102)
			}
			return true
		l101:
			position, tokenIndex, depth = position101, tokenIndex101, depth101
			return false
		},
		/* 32 tokdiv <- <('/' optspace)> */
		func() bool {
			position103, tokenIndex103, depth103 := position, tokenIndex, depth
			{
				position104 := position
				depth++
				if buffer[position] != '/' {
					goto l103
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l103
				}
				depth--
				add(Ruletokdiv, position104)
			}
			return true
		l103:
			position, tokenIndex, depth = position103, tokenIndex103, depth103
			return false
		},
		/* 33 endl <- <(optspace ';')> */
		func() bool {
			position105, tokenIndex105, depth105 := position, tokenIndex, depth
			{
				position106 := position
				depth++
				if !rules[Ruleoptspace]() {
					goto l105
				}
				if buffer[position] != ';' {
					goto l105
				}
				position++
				depth--
				add(Ruleendl, position106)
			}
			return true
		l105:
			position, tokenIndex, depth = position105, tokenIndex105, depth105
			return false
		},
		/* 34 rawmath <- <(value optspace (tokadd / toksub / tokmul / tokdiv) optspace value)> */
		func() bool {
			position107, tokenIndex107, depth107 := position, tokenIndex, depth
			{
				position108 := position
				depth++
				if !rules[Rulevalue]() {
					goto l107
				}
				if !rules[Ruleoptspace]() {
					goto l107
				}
				{
					position109, tokenIndex109, depth109 := position, tokenIndex, depth
					if !rules[Ruletokadd]() {
						goto l110
					}
					goto l109
				l110:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !rules[Ruletoksub]() {
						goto l111
					}
					goto l109
				l111:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !rules[Ruletokmul]() {
						goto l112
					}
					goto l109
				l112:
					position, tokenIndex, depth = position109, tokenIndex109, depth109
					if !rules[Ruletokdiv]() {
						goto l107
					}
				}
			l109:
				if !rules[Ruleoptspace]() {
					goto l107
				}
				if !rules[Rulevalue]() {
					goto l107
				}
				depth--
				add(Rulerawmath, position108)
			}
			return true
		l107:
			position, tokenIndex, depth = position107, tokenIndex107, depth107
			return false
		},
		/* 35 math <- <(optspace (rawmath / (popen rawmath pclose)) optspace)> */
		func() bool {
			position113, tokenIndex113, depth113 := position, tokenIndex, depth
			{
				position114 := position
				depth++
				if !rules[Ruleoptspace]() {
					goto l113
				}
				{
					position115, tokenIndex115, depth115 := position, tokenIndex, depth
					if !rules[Rulerawmath]() {
						goto l116
					}
					goto l115
				l116:
					position, tokenIndex, depth = position115, tokenIndex115, depth115
					if !rules[Rulepopen]() {
						goto l113
					}
					if !rules[Rulerawmath]() {
						goto l113
					}
					if !rules[Rulepclose]() {
						goto l113
					}
				}
			l115:
				if !rules[Ruleoptspace]() {
					goto l113
				}
				depth--
				add(Rulemath, position114)
			}
			return true
		l113:
			position, tokenIndex, depth = position113, tokenIndex113, depth113
			return false
		},
		/* 36 popen <- <('(' optspace)> */
		func() bool {
			position117, tokenIndex117, depth117 := position, tokenIndex, depth
			{
				position118 := position
				depth++
				if buffer[position] != '(' {
					goto l117
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l117
				}
				depth--
				add(Rulepopen, position118)
			}
			return true
		l117:
			position, tokenIndex, depth = position117, tokenIndex117, depth117
			return false
		},
		/* 37 pclose <- <(')' optspace)> */
		func() bool {
			position119, tokenIndex119, depth119 := position, tokenIndex, depth
			{
				position120 := position
				depth++
				if buffer[position] != ')' {
					goto l119
				}
				position++
				if !rules[Ruleoptspace]() {
					goto l119
				}
				depth--
				add(Rulepclose, position120)
			}
			return true
		l119:
			position, tokenIndex, depth = position119, tokenIndex119, depth119
			return false
		},
		/* 38 toklt <- <'<'> */
		func() bool {
			position121, tokenIndex121, depth121 := position, tokenIndex, depth
			{
				position122 := position
				depth++
				if buffer[position] != '<' {
					goto l121
				}
				position++
				depth--
				add(Ruletoklt, position122)
			}
			return true
		l121:
			position, tokenIndex, depth = position121, tokenIndex121, depth121
			return false
		},
		/* 39 tokgt <- <'>'> */
		func() bool {
			position123, tokenIndex123, depth123 := position, tokenIndex, depth
			{
				position124 := position
				depth++
				if buffer[position] != '>' {
					goto l123
				}
				position++
				depth--
				add(Ruletokgt, position124)
			}
			return true
		l123:
			position, tokenIndex, depth = position123, tokenIndex123, depth123
			return false
		},
		/* 40 tokeq <- <('=' '=')> */
		func() bool {
			position125, tokenIndex125, depth125 := position, tokenIndex, depth
			{
				position126 := position
				depth++
				if buffer[position] != '=' {
					goto l125
				}
				position++
				if buffer[position] != '=' {
					goto l125
				}
				position++
				depth--
				add(Ruletokeq, position126)
			}
			return true
		l125:
			position, tokenIndex, depth = position125, tokenIndex125, depth125
			return false
		},
		/* 41 tokle <- <('<' '=')> */
		func() bool {
			position127, tokenIndex127, depth127 := position, tokenIndex, depth
			{
				position128 := position
				depth++
				if buffer[position] != '<' {
					goto l127
				}
				position++
				if buffer[position] != '=' {
					goto l127
				}
				position++
				depth--
				add(Ruletokle, position128)
			}
			return true
		l127:
			position, tokenIndex, depth = position127, tokenIndex127, depth127
			return false
		},
		/* 42 tokge <- <('>' '=')> */
		func() bool {
			position129, tokenIndex129, depth129 := position, tokenIndex, depth
			{
				position130 := position
				depth++
				if buffer[position] != '>' {
					goto l129
				}
				position++
				if buffer[position] != '=' {
					goto l129
				}
				position++
				depth--
				add(Ruletokge, position130)
			}
			return true
		l129:
			position, tokenIndex, depth = position129, tokenIndex129, depth129
			return false
		},
		/* 43 tokne <- <('!' '=')> */
		func() bool {
			position131, tokenIndex131, depth131 := position, tokenIndex, depth
			{
				position132 := position
				depth++
				if buffer[position] != '!' {
					goto l131
				}
				position++
				if buffer[position] != '=' {
					goto l131
				}
				position++
				depth--
				add(Ruletokne, position132)
			}
			return true
		l131:
			position, tokenIndex, depth = position131, tokenIndex131, depth131
			return false
		},
		/* 44 comparisontoken <- <(toklt / tokgt / tokeq / tokle / tokge / tokne)> */
		func() bool {
			position133, tokenIndex133, depth133 := position, tokenIndex, depth
			{
				position134 := position
				depth++
				{
					position135, tokenIndex135, depth135 := position, tokenIndex, depth
					if !rules[Ruletoklt]() {
						goto l136
					}
					goto l135
				l136:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if !rules[Ruletokgt]() {
						goto l137
					}
					goto l135
				l137:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if !rules[Ruletokeq]() {
						goto l138
					}
					goto l135
				l138:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if !rules[Ruletokle]() {
						goto l139
					}
					goto l135
				l139:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if !rules[Ruletokge]() {
						goto l140
					}
					goto l135
				l140:
					position, tokenIndex, depth = position135, tokenIndex135, depth135
					if !rules[Ruletokne]() {
						goto l133
					}
				}
			l135:
				depth--
				add(Rulecomparisontoken, position134)
			}
			return true
		l133:
			position, tokenIndex, depth = position133, tokenIndex133, depth133
			return false
		},
		/* 45 comparison <- <(value minspace comparisontoken minspace value)> */
		func() bool {
			position141, tokenIndex141, depth141 := position, tokenIndex, depth
			{
				position142 := position
				depth++
				if !rules[Rulevalue]() {
					goto l141
				}
				if !rules[Ruleminspace]() {
					goto l141
				}
				if !rules[Rulecomparisontoken]() {
					goto l141
				}
				if !rules[Ruleminspace]() {
					goto l141
				}
				if !rules[Rulevalue]() {
					goto l141
				}
				depth--
				add(Rulecomparison, position142)
			}
			return true
		l141:
			position, tokenIndex, depth = position141, tokenIndex141, depth141
			return false
		},
		/* 46 comparison_paren <- <(popen optspace (comparison / value) optspace pclose)> */
		func() bool {
			position143, tokenIndex143, depth143 := position, tokenIndex, depth
			{
				position144 := position
				depth++
				if !rules[Rulepopen]() {
					goto l143
				}
				if !rules[Ruleoptspace]() {
					goto l143
				}
				{
					position145, tokenIndex145, depth145 := position, tokenIndex, depth
					if !rules[Rulecomparison]() {
						goto l146
					}
					goto l145
				l146:
					position, tokenIndex, depth = position145, tokenIndex145, depth145
					if !rules[Rulevalue]() {
						goto l143
					}
				}
			l145:
				if !rules[Ruleoptspace]() {
					goto l143
				}
				if !rules[Rulepclose]() {
					goto l143
				}
				depth--
				add(Rulecomparison_paren, position144)
			}
			return true
		l143:
			position, tokenIndex, depth = position143, tokenIndex143, depth143
			return false
		},
		/* 47 positivenum <- <([1-9] digit*)> */
		func() bool {
			position147, tokenIndex147, depth147 := position, tokenIndex, depth
			{
				position148 := position
				depth++
				if c := buffer[position]; c < '1' || c > '9' {
					goto l147
				}
				position++
			l149:
				{
					position150, tokenIndex150, depth150 := position, tokenIndex, depth
					if !rules[Ruledigit]() {
						goto l150
					}
					goto l149
				l150:
					position, tokenIndex, depth = position150, tokenIndex150, depth150
				}
				depth--
				add(Rulepositivenum, position148)
			}
			return true
		l147:
			position, tokenIndex, depth = position147, tokenIndex147, depth147
			return false
		},
		/* 48 negativenum <- <('-' positivenum)> */
		func() bool {
			position151, tokenIndex151, depth151 := position, tokenIndex, depth
			{
				position152 := position
				depth++
				if buffer[position] != '-' {
					goto l151
				}
				position++
				if !rules[Rulepositivenum]() {
					goto l151
				}
				depth--
				add(Rulenegativenum, position152)
			}
			return true
		l151:
			position, tokenIndex, depth = position151, tokenIndex151, depth151
			return false
		},
		/* 49 number <- <(positivenum / negativenum)> */
		func() bool {
			position153, tokenIndex153, depth153 := position, tokenIndex, depth
			{
				position154 := position
				depth++
				{
					position155, tokenIndex155, depth155 := position, tokenIndex, depth
					if !rules[Rulepositivenum]() {
						goto l156
					}
					goto l155
				l156:
					position, tokenIndex, depth = position155, tokenIndex155, depth155
					if !rules[Rulenegativenum]() {
						goto l153
					}
				}
			l155:
				depth--
				add(Rulenumber, position154)
			}
			return true
		l153:
			position, tokenIndex, depth = position153, tokenIndex153, depth153
			return false
		},
		/* 50 digit <- <[0-9]> */
		func() bool {
			position157, tokenIndex157, depth157 := position, tokenIndex, depth
			{
				position158 := position
				depth++
				if c := buffer[position]; c < '0' || c > '9' {
					goto l157
				}
				position++
				depth--
				add(Ruledigit, position158)
			}
			return true
		l157:
			position, tokenIndex, depth = position157, tokenIndex157, depth157
			return false
		},
		/* 51 commentblock <- <('/' '*' (!('*' '/') .)* ('*' '/'))> */
		func() bool {
			position159, tokenIndex159, depth159 := position, tokenIndex, depth
			{
				position160 := position
				depth++
				if buffer[position] != '/' {
					goto l159
				}
				position++
				if buffer[position] != '*' {
					goto l159
				}
				position++
			l161:
				{
					position162, tokenIndex162, depth162 := position, tokenIndex, depth
					{
						position163, tokenIndex163, depth163 := position, tokenIndex, depth
						if buffer[position] != '*' {
							goto l163
						}
						position++
						if buffer[position] != '/' {
							goto l163
						}
						position++
						goto l162
					l163:
						position, tokenIndex, depth = position163, tokenIndex163, depth163
					}
					if !matchDot() {
						goto l162
					}
					goto l161
				l162:
					position, tokenIndex, depth = position162, tokenIndex162, depth162
				}
				if buffer[position] != '*' {
					goto l159
				}
				position++
				if buffer[position] != '/' {
					goto l159
				}
				position++
				depth--
				add(Rulecommentblock, position160)
			}
			return true
		l159:
			position, tokenIndex, depth = position159, tokenIndex159, depth159
			return false
		},
		/* 52 commentdoubleslash <- <('/' '/' (!('\n' / '\r') .)* space)> */
		func() bool {
			position164, tokenIndex164, depth164 := position, tokenIndex, depth
			{
				position165 := position
				depth++
				if buffer[position] != '/' {
					goto l164
				}
				position++
				if buffer[position] != '/' {
					goto l164
				}
				position++
			l166:
				{
					position167, tokenIndex167, depth167 := position, tokenIndex, depth
					{
						position168, tokenIndex168, depth168 := position, tokenIndex, depth
						{
							position169, tokenIndex169, depth169 := position, tokenIndex, depth
							if buffer[position] != '\n' {
								goto l170
							}
							position++
							goto l169
						l170:
							position, tokenIndex, depth = position169, tokenIndex169, depth169
							if buffer[position] != '\r' {
								goto l168
							}
							position++
						}
					l169:
						goto l167
					l168:
						position, tokenIndex, depth = position168, tokenIndex168, depth168
					}
					if !matchDot() {
						goto l167
					}
					goto l166
				l167:
					position, tokenIndex, depth = position167, tokenIndex167, depth167
				}
				if !rules[Rulespace]() {
					goto l164
				}
				depth--
				add(Rulecommentdoubleslash, position165)
			}
			return true
		l164:
			position, tokenIndex, depth = position164, tokenIndex164, depth164
			return false
		},
		/* 53 comment <- <(commentblock / commentdoubleslash)> */
		func() bool {
			position171, tokenIndex171, depth171 := position, tokenIndex, depth
			{
				position172 := position
				depth++
				{
					position173, tokenIndex173, depth173 := position, tokenIndex, depth
					if !rules[Rulecommentblock]() {
						goto l174
					}
					goto l173
				l174:
					position, tokenIndex, depth = position173, tokenIndex173, depth173
					if !rules[Rulecommentdoubleslash]() {
						goto l171
					}
				}
			l173:
				depth--
				add(Rulecomment, position172)
			}
			return true
		l171:
			position, tokenIndex, depth = position171, tokenIndex171, depth171
			return false
		},
		/* 54 literalspace <- <(' ' / '\t' / '\n' / '\r')> */
		func() bool {
			position175, tokenIndex175, depth175 := position, tokenIndex, depth
			{
				position176 := position
				depth++
				{
					position177, tokenIndex177, depth177 := position, tokenIndex, depth
					if buffer[position] != ' ' {
						goto l178
					}
					position++
					goto l177
				l178:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if buffer[position] != '\t' {
						goto l179
					}
					position++
					goto l177
				l179:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if buffer[position] != '\n' {
						goto l180
					}
					position++
					goto l177
				l180:
					position, tokenIndex, depth = position177, tokenIndex177, depth177
					if buffer[position] != '\r' {
						goto l175
					}
					position++
				}
			l177:
				depth--
				add(Ruleliteralspace, position176)
			}
			return true
		l175:
			position, tokenIndex, depth = position175, tokenIndex175, depth175
			return false
		},
		/* 55 space <- <(comment / literalspace)> */
		func() bool {
			position181, tokenIndex181, depth181 := position, tokenIndex, depth
			{
				position182 := position
				depth++
				{
					position183, tokenIndex183, depth183 := position, tokenIndex, depth
					if !rules[Rulecomment]() {
						goto l184
					}
					goto l183
				l184:
					position, tokenIndex, depth = position183, tokenIndex183, depth183
					if !rules[Ruleliteralspace]() {
						goto l181
					}
				}
			l183:
				depth--
				add(Rulespace, position182)
			}
			return true
		l181:
			position, tokenIndex, depth = position181, tokenIndex181, depth181
			return false
		},
		/* 56 minspace <- <space+> */
		func() bool {
			position185, tokenIndex185, depth185 := position, tokenIndex, depth
			{
				position186 := position
				depth++
				if !rules[Rulespace]() {
					goto l185
				}
			l187:
				{
					position188, tokenIndex188, depth188 := position, tokenIndex, depth
					if !rules[Rulespace]() {
						goto l188
					}
					goto l187
				l188:
					position, tokenIndex, depth = position188, tokenIndex188, depth188
				}
				depth--
				add(Ruleminspace, position186)
			}
			return true
		l185:
			position, tokenIndex, depth = position185, tokenIndex185, depth185
			return false
		},
		/* 57 optspace <- <space*> */
		func() bool {
			{
				position190 := position
				depth++
			l191:
				{
					position192, tokenIndex192, depth192 := position, tokenIndex, depth
					if !rules[Rulespace]() {
						goto l192
					}
					goto l191
				l192:
					position, tokenIndex, depth = position192, tokenIndex192, depth192
				}
				depth--
				add(Ruleoptspace, position190)
			}
			return true
		},
	}
	p.rules = rules
}
