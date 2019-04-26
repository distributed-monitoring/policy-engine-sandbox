package parser

import (
	"fmt"
	"os"
)

type ExprTypes int

const (
	ExprNone ExprTypes = iota
	ExprNum
	ExprVar
	ExprStr
	ExprEq
	ExprNe
	ExprLe
	ExprGe
	ExprLt
	ExprGt
)

type ExprSymbol struct {
	Types   ExprTypes
	ExprNum string
	ExprVar string
	ExprStr string
}

type ExprCond struct {
	Left  *ExprSymbol
	Ops   ExprTypes
	Right *ExprSymbol
}

type PolicyExpr struct {
	Left  *ExprSymbol
	Ops   ExprTypes
	Right *ExprSymbol
}

func (p *PolicyExpr) AddOps(ops ExprTypes) {
	if p.Ops == ExprNone {
		p.Ops = ops
	} else {
		fmt.Fprintf(os.Stderr, "error")
	}
}

func (p *PolicyExpr) AddNum(s string) {
	if p.Left == nil {
		p.Left = &ExprSymbol{
			Types:   ExprNum,
			ExprNum: s,
		}
	} else if p.Right == nil {
		p.Right = &ExprSymbol{
			Types:   ExprNum,
			ExprNum: s,
		}
	} else {
		fmt.Fprintf(os.Stderr, "error")
	}
}

func (p *PolicyExpr) AddVar(s string) {
	if p.Left == nil {
		p.Left = &ExprSymbol{
			Types:   ExprVar,
			ExprVar: s,
		}
	} else if p.Right == nil {
		p.Right = &ExprSymbol{
			Types:   ExprVar,
			ExprVar: s,
		}
	} else {
		fmt.Fprintf(os.Stderr, "error")
	}
}

func (p *PolicyExpr) AddStr(s string) {
	if p.Left == nil {
		p.Left = &ExprSymbol{
			Types:   ExprStr,
			ExprNum: s,
		}
	} else if p.Right == nil {
		p.Right = &ExprSymbol{
			Types:   ExprStr,
			ExprNum: s,
		}
	} else {
		fmt.Fprintf(os.Stderr, "error")
	}
}

func (s *ExprSymbol) Print() {
	switch s.Types {
	case ExprNum:
		fmt.Printf("%s", s.ExprNum)
	case ExprVar:
		fmt.Printf("%s", s.ExprVar)
	case ExprStr:
		fmt.Printf("'%s'", s.ExprStr)
	default:
		fmt.Printf("??%d??", s.Types)
	}
}

func (t *ExprTypes) Print() {
	ops_symbol := ""
	switch *t {
	case ExprEq:
		ops_symbol = "=="
	case ExprNe:
		ops_symbol = "!="
	case ExprLe:
		ops_symbol = "<="
	case ExprGe:
		ops_symbol = ">="
	case ExprLt:
		ops_symbol = "<"
	case ExprGt:
		ops_symbol = ">"
	}
	fmt.Printf(" %s ", ops_symbol)
}

func (policy *PolicyExpr) PrintPolicy() {
	policy.Left.Print()
	policy.Ops.Print()
	policy.Right.Print()
}

func Policyexpr_main(expr_val string) *Parser {
	fmt.Printf("Text of expr: %s\n", expr_val)
	parser := &Parser{Buffer: expr_val}
	parser.Init()

	err := parser.Parse()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		return nil
	}

	parser.Execute()
	parser.PrintPolicy()
	fmt.Printf("\ndone!\n")
	return parser
}
