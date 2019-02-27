package policyexpr

//go:generate peg policy.peg

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
)

const endSymbol rune = 1114112

/* The rule types inferred from the grammar are below. */
type pegRule uint8

const (
	ruleUnknown pegRule = iota
	ruleroot
	ruleexpression
	rulecondition
	rulesymbol
	rulenumbers
	rulevariables
	rulestrings
	ruleStringChar
	ruleidchar
	ruleops
	ruleopeq
	ruleopne
	ruleople
	ruleopge
	ruleoplt
	ruleopgt
	rulesp
	rulePegText
	ruleAction0
	ruleAction1
	ruleAction2
	ruleAction3
	ruleAction4
	ruleAction5
	ruleAction6
	ruleAction7
	ruleAction8
)

var rul3s = [...]string{
	"Unknown",
	"root",
	"expression",
	"condition",
	"symbol",
	"numbers",
	"variables",
	"strings",
	"StringChar",
	"idchar",
	"ops",
	"opeq",
	"opne",
	"ople",
	"opge",
	"oplt",
	"opgt",
	"sp",
	"PegText",
	"Action0",
	"Action1",
	"Action2",
	"Action3",
	"Action4",
	"Action5",
	"Action6",
	"Action7",
	"Action8",
}

type token32 struct {
	pegRule
	begin, end uint32
}

func (t *token32) String() string {
	return fmt.Sprintf("\x1B[34m%v\x1B[m %v %v", rul3s[t.pegRule], t.begin, t.end)
}

type node32 struct {
	token32
	up, next *node32
}

func (node *node32) print(w io.Writer, pretty bool, buffer string) {
	var print func(node *node32, depth int)
	print = func(node *node32, depth int) {
		for node != nil {
			for c := 0; c < depth; c++ {
				fmt.Fprintf(w, " ")
			}
			rule := rul3s[node.pegRule]
			quote := strconv.Quote(string(([]rune(buffer)[node.begin:node.end])))
			if !pretty {
				fmt.Fprintf(w, "%v %v\n", rule, quote)
			} else {
				fmt.Fprintf(w, "\x1B[34m%v\x1B[m %v\n", rule, quote)
			}
			if node.up != nil {
				print(node.up, depth+1)
			}
			node = node.next
		}
	}
	print(node, 0)
}

func (node *node32) Print(w io.Writer, buffer string) {
	node.print(w, false, buffer)
}

func (node *node32) PrettyPrint(w io.Writer, buffer string) {
	node.print(w, true, buffer)
}

type tokens32 struct {
	tree []token32
}

func (t *tokens32) Trim(length uint32) {
	t.tree = t.tree[:length]
}

func (t *tokens32) Print() {
	for _, token := range t.tree {
		fmt.Println(token.String())
	}
}

func (t *tokens32) AST() *node32 {
	type element struct {
		node *node32
		down *element
	}
	tokens := t.Tokens()
	var stack *element
	for _, token := range tokens {
		if token.begin == token.end {
			continue
		}
		node := &node32{token32: token}
		for stack != nil && stack.node.begin >= token.begin && stack.node.end <= token.end {
			stack.node.next = node.up
			node.up = stack.node
			stack = stack.down
		}
		stack = &element{node: node, down: stack}
	}
	if stack != nil {
		return stack.node
	}
	return nil
}

func (t *tokens32) PrintSyntaxTree(buffer string) {
	t.AST().Print(os.Stdout, buffer)
}

func (t *tokens32) WriteSyntaxTree(w io.Writer, buffer string) {
	t.AST().Print(w, buffer)
}

func (t *tokens32) PrettyPrintSyntaxTree(buffer string) {
	t.AST().PrettyPrint(os.Stdout, buffer)
}

func (t *tokens32) Add(rule pegRule, begin, end, index uint32) {
	if tree := t.tree; int(index) >= len(tree) {
		expanded := make([]token32, 2*len(tree))
		copy(expanded, tree)
		t.tree = expanded
	}
	t.tree[index] = token32{
		pegRule: rule,
		begin:   begin,
		end:     end,
	}
}

func (t *tokens32) Tokens() []token32 {
	return t.tree
}

type Parser struct {
	PolicyExpr

	Buffer string
	buffer []rune
	rules  [28]func() bool
	parse  func(rule ...int) error
	reset  func()
	Pretty bool
	tokens32
}

func (p *Parser) Parse(rule ...int) error {
	return p.parse(rule...)
}

func (p *Parser) Reset() {
	p.reset()
}

type textPosition struct {
	line, symbol int
}

type textPositionMap map[int]textPosition

func translatePositions(buffer []rune, positions []int) textPositionMap {
	length, translations, j, line, symbol := len(positions), make(textPositionMap, len(positions)), 0, 1, 0
	sort.Ints(positions)

search:
	for i, c := range buffer {
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
	p   *Parser
	max token32
}

func (e *parseError) Error() string {
	tokens, err := []token32{e.max}, "\n"
	positions, p := make([]int, 2*len(tokens)), 0
	for _, token := range tokens {
		positions[p], p = int(token.begin), p+1
		positions[p], p = int(token.end), p+1
	}
	translations := translatePositions(e.p.buffer, positions)
	format := "parse error near %v (line %v symbol %v - line %v symbol %v):\n%v\n"
	if e.p.Pretty {
		format = "parse error near \x1B[34m%v\x1B[m (line %v symbol %v - line %v symbol %v):\n%v\n"
	}
	for _, token := range tokens {
		begin, end := int(token.begin), int(token.end)
		err += fmt.Sprintf(format,
			rul3s[token.pegRule],
			translations[begin].line, translations[begin].symbol,
			translations[end].line, translations[end].symbol,
			strconv.Quote(string(e.p.buffer[begin:end])))
	}

	return err
}

func (p *Parser) PrintSyntaxTree() {
	if p.Pretty {
		p.tokens32.PrettyPrintSyntaxTree(p.Buffer)
	} else {
		p.tokens32.PrintSyntaxTree(p.Buffer)
	}
}

func (p *Parser) WriteSyntaxTree(w io.Writer) {
	p.tokens32.WriteSyntaxTree(w, p.Buffer)
}

func (p *Parser) Execute() {
	buffer, _buffer, text, begin, end := p.Buffer, p.buffer, "", 0, 0
	for _, token := range p.Tokens() {
		switch token.pegRule {

		case rulePegText:
			begin, end = int(token.begin), int(token.end)
			text = string(_buffer[begin:end])

		case ruleAction0:
			p.AddNum(buffer[begin:end])
		case ruleAction1:
			p.AddVar(buffer[begin:end])
		case ruleAction2:
			p.AddStr(buffer[begin:end])
		case ruleAction3:
			p.AddOps(ExprEq)
		case ruleAction4:
			p.AddOps(ExprNe)
		case ruleAction5:
			p.AddOps(ExprLe)
		case ruleAction6:
			p.AddOps(ExprGe)
		case ruleAction7:
			p.AddOps(ExprLt)
		case ruleAction8:
			p.AddOps(ExprGt)

		}
	}
	_, _, _, _, _ = buffer, _buffer, text, begin, end
}

func (p *Parser) Init() {
	var (
		max                  token32
		position, tokenIndex uint32
		buffer               []rune
	)
	p.reset = func() {
		max = token32{}
		position, tokenIndex = 0, 0

		p.buffer = []rune(p.Buffer)
		if len(p.buffer) == 0 || p.buffer[len(p.buffer)-1] != endSymbol {
			p.buffer = append(p.buffer, endSymbol)
		}
		buffer = p.buffer
	}
	p.reset()

	_rules := p.rules
	tree := tokens32{tree: make([]token32, math.MaxInt16)}
	p.parse = func(rule ...int) error {
		r := 1
		if len(rule) > 0 {
			r = rule[0]
		}
		matches := p.rules[r]()
		p.tokens32 = tree
		if matches {
			p.Trim(tokenIndex)
			return nil
		}
		return &parseError{p, max}
	}

	add := func(rule pegRule, begin uint32) {
		tree.Add(rule, begin, position, tokenIndex)
		tokenIndex++
		if begin != position && position > max.end {
			max = token32{rule, begin, position}
		}
	}

	matchDot := func() bool {
		if buffer[position] != endSymbol {
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

	_rules = [...]func() bool{
		nil,
		/* 0 root <- <(sp expression !.)> */
		func() bool {
			position0, tokenIndex0 := position, tokenIndex
			{
				position1 := position
				if !_rules[rulesp]() {
					goto l0
				}
				if !_rules[ruleexpression]() {
					goto l0
				}
				{
					position2, tokenIndex2 := position, tokenIndex
					if !matchDot() {
						goto l2
					}
					goto l0
				l2:
					position, tokenIndex = position2, tokenIndex2
				}
				add(ruleroot, position1)
			}
			return true
		l0:
			position, tokenIndex = position0, tokenIndex0
			return false
		},
		/* 1 expression <- <condition> */
		func() bool {
			position3, tokenIndex3 := position, tokenIndex
			{
				position4 := position
				if !_rules[rulecondition]() {
					goto l3
				}
				add(ruleexpression, position4)
			}
			return true
		l3:
			position, tokenIndex = position3, tokenIndex3
			return false
		},
		/* 2 condition <- <(symbol ops symbol)> */
		func() bool {
			position5, tokenIndex5 := position, tokenIndex
			{
				position6 := position
				if !_rules[rulesymbol]() {
					goto l5
				}
				if !_rules[ruleops]() {
					goto l5
				}
				if !_rules[rulesymbol]() {
					goto l5
				}
				add(rulecondition, position6)
			}
			return true
		l5:
			position, tokenIndex = position5, tokenIndex5
			return false
		},
		/* 3 symbol <- <((numbers sp) / (strings sp) / (variables sp))> */
		func() bool {
			position7, tokenIndex7 := position, tokenIndex
			{
				position8 := position
				{
					position9, tokenIndex9 := position, tokenIndex
					if !_rules[rulenumbers]() {
						goto l10
					}
					if !_rules[rulesp]() {
						goto l10
					}
					goto l9
				l10:
					position, tokenIndex = position9, tokenIndex9
					if !_rules[rulestrings]() {
						goto l11
					}
					if !_rules[rulesp]() {
						goto l11
					}
					goto l9
				l11:
					position, tokenIndex = position9, tokenIndex9
					if !_rules[rulevariables]() {
						goto l7
					}
					if !_rules[rulesp]() {
						goto l7
					}
				}
			l9:
				add(rulesymbol, position8)
			}
			return true
		l7:
			position, tokenIndex = position7, tokenIndex7
			return false
		},
		/* 4 numbers <- <(<[0-9]+> Action0)> */
		func() bool {
			position12, tokenIndex12 := position, tokenIndex
			{
				position13 := position
				{
					position14 := position
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l12
					}
					position++
				l15:
					{
						position16, tokenIndex16 := position, tokenIndex
						if c := buffer[position]; c < rune('0') || c > rune('9') {
							goto l16
						}
						position++
						goto l15
					l16:
						position, tokenIndex = position16, tokenIndex16
					}
					add(rulePegText, position14)
				}
				if !_rules[ruleAction0]() {
					goto l12
				}
				add(rulenumbers, position13)
			}
			return true
		l12:
			position, tokenIndex = position12, tokenIndex12
			return false
		},
		/* 5 variables <- <(<idchar*> Action1)> */
		func() bool {
			position17, tokenIndex17 := position, tokenIndex
			{
				position18 := position
				{
					position19 := position
				l20:
					{
						position21, tokenIndex21 := position, tokenIndex
						if !_rules[ruleidchar]() {
							goto l21
						}
						goto l20
					l21:
						position, tokenIndex = position21, tokenIndex21
					}
					add(rulePegText, position19)
				}
				if !_rules[ruleAction1]() {
					goto l17
				}
				add(rulevariables, position18)
			}
			return true
		l17:
			position, tokenIndex = position17, tokenIndex17
			return false
		},
		/* 6 strings <- <('"' <StringChar*> '"' sp Action2)> */
		func() bool {
			position22, tokenIndex22 := position, tokenIndex
			{
				position23 := position
				if buffer[position] != rune('"') {
					goto l22
				}
				position++
				{
					position24 := position
				l25:
					{
						position26, tokenIndex26 := position, tokenIndex
						if !_rules[ruleStringChar]() {
							goto l26
						}
						goto l25
					l26:
						position, tokenIndex = position26, tokenIndex26
					}
					add(rulePegText, position24)
				}
				if buffer[position] != rune('"') {
					goto l22
				}
				position++
				if !_rules[rulesp]() {
					goto l22
				}
				if !_rules[ruleAction2]() {
					goto l22
				}
				add(rulestrings, position23)
			}
			return true
		l22:
			position, tokenIndex = position22, tokenIndex22
			return false
		},
		/* 7 StringChar <- <(!('"' / '\n' / '\\') .)> */
		func() bool {
			position27, tokenIndex27 := position, tokenIndex
			{
				position28 := position
				{
					position29, tokenIndex29 := position, tokenIndex
					{
						position30, tokenIndex30 := position, tokenIndex
						if buffer[position] != rune('"') {
							goto l31
						}
						position++
						goto l30
					l31:
						position, tokenIndex = position30, tokenIndex30
						if buffer[position] != rune('\n') {
							goto l32
						}
						position++
						goto l30
					l32:
						position, tokenIndex = position30, tokenIndex30
						if buffer[position] != rune('\\') {
							goto l29
						}
						position++
					}
				l30:
					goto l27
				l29:
					position, tokenIndex = position29, tokenIndex29
				}
				if !matchDot() {
					goto l27
				}
				add(ruleStringChar, position28)
			}
			return true
		l27:
			position, tokenIndex = position27, tokenIndex27
			return false
		},
		/* 8 idchar <- <([a-z] / [A-Z] / [0-9] / '_' / '.' / '-')> */
		func() bool {
			position33, tokenIndex33 := position, tokenIndex
			{
				position34 := position
				{
					position35, tokenIndex35 := position, tokenIndex
					if c := buffer[position]; c < rune('a') || c > rune('z') {
						goto l36
					}
					position++
					goto l35
				l36:
					position, tokenIndex = position35, tokenIndex35
					if c := buffer[position]; c < rune('A') || c > rune('Z') {
						goto l37
					}
					position++
					goto l35
				l37:
					position, tokenIndex = position35, tokenIndex35
					if c := buffer[position]; c < rune('0') || c > rune('9') {
						goto l38
					}
					position++
					goto l35
				l38:
					position, tokenIndex = position35, tokenIndex35
					if buffer[position] != rune('_') {
						goto l39
					}
					position++
					goto l35
				l39:
					position, tokenIndex = position35, tokenIndex35
					if buffer[position] != rune('.') {
						goto l40
					}
					position++
					goto l35
				l40:
					position, tokenIndex = position35, tokenIndex35
					if buffer[position] != rune('-') {
						goto l33
					}
					position++
				}
			l35:
				add(ruleidchar, position34)
			}
			return true
		l33:
			position, tokenIndex = position33, tokenIndex33
			return false
		},
		/* 9 ops <- <((opeq sp Action3) / (opne sp Action4) / (ople sp Action5) / (opge sp Action6) / (oplt sp Action7) / (opgt sp Action8))> */
		func() bool {
			position41, tokenIndex41 := position, tokenIndex
			{
				position42 := position
				{
					position43, tokenIndex43 := position, tokenIndex
					if !_rules[ruleopeq]() {
						goto l44
					}
					if !_rules[rulesp]() {
						goto l44
					}
					if !_rules[ruleAction3]() {
						goto l44
					}
					goto l43
				l44:
					position, tokenIndex = position43, tokenIndex43
					if !_rules[ruleopne]() {
						goto l45
					}
					if !_rules[rulesp]() {
						goto l45
					}
					if !_rules[ruleAction4]() {
						goto l45
					}
					goto l43
				l45:
					position, tokenIndex = position43, tokenIndex43
					if !_rules[ruleople]() {
						goto l46
					}
					if !_rules[rulesp]() {
						goto l46
					}
					if !_rules[ruleAction5]() {
						goto l46
					}
					goto l43
				l46:
					position, tokenIndex = position43, tokenIndex43
					if !_rules[ruleopge]() {
						goto l47
					}
					if !_rules[rulesp]() {
						goto l47
					}
					if !_rules[ruleAction6]() {
						goto l47
					}
					goto l43
				l47:
					position, tokenIndex = position43, tokenIndex43
					if !_rules[ruleoplt]() {
						goto l48
					}
					if !_rules[rulesp]() {
						goto l48
					}
					if !_rules[ruleAction7]() {
						goto l48
					}
					goto l43
				l48:
					position, tokenIndex = position43, tokenIndex43
					if !_rules[ruleopgt]() {
						goto l41
					}
					if !_rules[rulesp]() {
						goto l41
					}
					if !_rules[ruleAction8]() {
						goto l41
					}
				}
			l43:
				add(ruleops, position42)
			}
			return true
		l41:
			position, tokenIndex = position41, tokenIndex41
			return false
		},
		/* 10 opeq <- <('=' '=')> */
		func() bool {
			position49, tokenIndex49 := position, tokenIndex
			{
				position50 := position
				if buffer[position] != rune('=') {
					goto l49
				}
				position++
				if buffer[position] != rune('=') {
					goto l49
				}
				position++
				add(ruleopeq, position50)
			}
			return true
		l49:
			position, tokenIndex = position49, tokenIndex49
			return false
		},
		/* 11 opne <- <('!' '=')> */
		func() bool {
			position51, tokenIndex51 := position, tokenIndex
			{
				position52 := position
				if buffer[position] != rune('!') {
					goto l51
				}
				position++
				if buffer[position] != rune('=') {
					goto l51
				}
				position++
				add(ruleopne, position52)
			}
			return true
		l51:
			position, tokenIndex = position51, tokenIndex51
			return false
		},
		/* 12 ople <- <('<' '=')> */
		func() bool {
			position53, tokenIndex53 := position, tokenIndex
			{
				position54 := position
				if buffer[position] != rune('<') {
					goto l53
				}
				position++
				if buffer[position] != rune('=') {
					goto l53
				}
				position++
				add(ruleople, position54)
			}
			return true
		l53:
			position, tokenIndex = position53, tokenIndex53
			return false
		},
		/* 13 opge <- <'='> */
		func() bool {
			position55, tokenIndex55 := position, tokenIndex
			{
				position56 := position
				if buffer[position] != rune('=') {
					goto l55
				}
				position++
				add(ruleopge, position56)
			}
			return true
		l55:
			position, tokenIndex = position55, tokenIndex55
			return false
		},
		/* 14 oplt <- <'<'> */
		func() bool {
			position57, tokenIndex57 := position, tokenIndex
			{
				position58 := position
				if buffer[position] != rune('<') {
					goto l57
				}
				position++
				add(ruleoplt, position58)
			}
			return true
		l57:
			position, tokenIndex = position57, tokenIndex57
			return false
		},
		/* 15 opgt <- <'>'> */
		func() bool {
			position59, tokenIndex59 := position, tokenIndex
			{
				position60 := position
				if buffer[position] != rune('>') {
					goto l59
				}
				position++
				add(ruleopgt, position60)
			}
			return true
		l59:
			position, tokenIndex = position59, tokenIndex59
			return false
		},
		/* 16 sp <- <(' ' / '\t')*> */
		func() bool {
			{
				position62 := position
			l63:
				{
					position64, tokenIndex64 := position, tokenIndex
					{
						position65, tokenIndex65 := position, tokenIndex
						if buffer[position] != rune(' ') {
							goto l66
						}
						position++
						goto l65
					l66:
						position, tokenIndex = position65, tokenIndex65
						if buffer[position] != rune('\t') {
							goto l64
						}
						position++
					}
				l65:
					goto l63
				l64:
					position, tokenIndex = position64, tokenIndex64
				}
				add(rulesp, position62)
			}
			return true
		},
		nil,
		/* 19 Action0 <- <{ p.AddNum(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction0, position)
			}
			return true
		},
		/* 20 Action1 <- <{ p.AddVar(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction1, position)
			}
			return true
		},
		/* 21 Action2 <- <{ p.AddStr(buffer[begin:end]) }> */
		func() bool {
			{
				add(ruleAction2, position)
			}
			return true
		},
		/* 22 Action3 <- <{ p.AddOps(ExprEq) }> */
		func() bool {
			{
				add(ruleAction3, position)
			}
			return true
		},
		/* 23 Action4 <- <{ p.AddOps(ExprNe) }> */
		func() bool {
			{
				add(ruleAction4, position)
			}
			return true
		},
		/* 24 Action5 <- <{ p.AddOps(ExprLe) }> */
		func() bool {
			{
				add(ruleAction5, position)
			}
			return true
		},
		/* 25 Action6 <- <{ p.AddOps(ExprGe) }> */
		func() bool {
			{
				add(ruleAction6, position)
			}
			return true
		},
		/* 26 Action7 <- <{ p.AddOps(ExprLt) }> */
		func() bool {
			{
				add(ruleAction7, position)
			}
			return true
		},
		/* 27 Action8 <- <{ p.AddOps(ExprGt) }> */
		func() bool {
			{
				add(ruleAction8, position)
			}
			return true
		},
	}
	p.rules = _rules
}
