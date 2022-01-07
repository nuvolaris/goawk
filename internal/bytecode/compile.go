package bytecode

import (
	"fmt"
	"regexp"

	"github.com/benhoyt/goawk/internal/ast"
	"github.com/benhoyt/goawk/lexer"
	"github.com/benhoyt/goawk/parser"
)

type Program struct {
	Begin       []Opcode
	Actions     []Action
	End         []Opcode
	Functions   []Function
	ScalarNames []string
	ArrayNames  []string
	Nums        []float64
	Strs        []string
	Regexes     []*regexp.Regexp
}

type Action struct {
	Pattern []Opcode
	Body    []Opcode
}

type Function struct {
	Name   string
	Params []string
	Arrays []bool
	Body   []Opcode
}

func Compile(prog *parser.Program) *Program {
	p := &Program{}
	c := &compiler{}

	for _, stmts := range prog.Begin {
		p.Begin = append(p.Begin, c.stmts(stmts)...)
	}
	//for _, action := range prog.Actions {
	//}
	for _, stmts := range prog.End {
		p.End = append(p.End, c.stmts(stmts)...)
	}

	p.ScalarNames = make([]string, len(prog.Scalars))
	for name, index := range prog.Scalars {
		p.ScalarNames[index] = name
	}
	p.ArrayNames = make([]string, len(prog.Arrays))
	for name, index := range prog.Arrays {
		p.ArrayNames[index] = name
	}
	p.Nums = c.nums
	p.Strs = c.strs
	p.Regexes = c.regexes
	return p
}

type compiler struct {
	nums    []float64
	strs    []string
	regexes []*regexp.Regexp
}

func (c *compiler) stmts(stmts []ast.Stmt) []Opcode {
	var code []Opcode
	for _, stmt := range stmts {
		code = append(code, c.stmt(stmt)...)
	}
	return code
}

func (c *compiler) stmt(stmt ast.Stmt) []Opcode {
	var code []Opcode
	switch s := stmt.(type) {
	case *ast.ExprStmt:
		// Optimize assignment expressions to avoid Dupe and Drop
		switch expr := s.Expr.(type) {
		case *ast.AssignExpr:
			switch left := expr.Left.(type) {
			case *ast.VarExpr:
				if left.Scope == ast.ScopeGlobal {
					if left.Index > 255 {
						panic("TODO: ExprStmt assign index too big")
					}
					code = append(code, c.expr(expr.Right)...)
					code = append(code, AssignGlobal, Opcode(left.Index))
					return code
				}
			}
		case *ast.IncrExpr:
			if !expr.Pre {
				switch target := expr.Expr.(type) {
				case *ast.VarExpr:
					if target.Scope == ast.ScopeGlobal {
						if target.Index > 255 {
							panic("TODO: ExprStmt incr index too big")
						}
						code = append(code, PostIncrGlobal, Opcode(target.Index))
						return code
					}
				}
			}
		case *ast.AugAssignExpr:
			switch left := expr.Left.(type) {
			case *ast.VarExpr:
				if left.Scope == ast.ScopeGlobal {
					if left.Index > 255 {
						panic("TODO: ExprStmt aug assign index too big")
					}
					code = append(code, c.expr(expr.Right)...)
					code = append(code, AugAssignGlobal, Opcode(expr.Op), Opcode(left.Index))
					return code
				}
			}
		}
		code = append(code, c.expr(s.Expr)...)
		code = append(code, Drop)

	case *ast.PrintStmt:
		if len(s.Args) > 255 {
			panic("TODO: too many args to print")
		}
		for _, a := range s.Args {
			code = append(code, c.expr(a)...)
		}
		if s.Redirect == lexer.ILLEGAL {
			code = append(code, Print, Opcode(len(s.Args)))
		} else {
			code = append(code, c.expr(s.Dest)...)
			code = append(code, PrintRedirect, Opcode(len(s.Args)), Opcode(s.Redirect))
		}

	//case *ast.PrintfStmt:
	//
	//case *ast.IfStmt:

	case *ast.ForStmt:
		if s.Pre != nil {
			code = append(code, c.stmt(s.Pre)...)
		}
		// Optimization: include condition once before loop and at the end
		var forwardMark int
		if s.Cond != nil {
			code = append(code, c.expr(s.Cond)...)
			forwardMark = len(code)
			code = append(code, JumpFalse, 0)
		}

		loopStart := len(code)
		code = append(code, c.stmts(s.Body)...)
		if s.Post != nil {
			code = append(code, c.stmt(s.Post)...)
		}

		if s.Cond != nil {
			// TODO: if s.Cond is BinaryExpr num == != < > <= >= or str == != then use JumpLess and similar optimizations

			done := false
			switch cond := s.Cond.(type) {
			case *ast.BinaryExpr:
				switch cond.Op {
				case lexer.LESS:
					if _, ok := cond.Right.(*ast.NumExpr); ok {
						done = true
						code = append(code, c.expr(cond.Left)...)
						code = append(code, c.expr(cond.Right)...)
						offset := loopStart - (len(code) + 2)
						if offset > 255 {
							panic("TODO: for jump offset too big")
						}
						code = append(code, JumpNumLess, Opcode(int8(offset)))
					}
				}
			}
			if !done {
				code = append(code, c.expr(s.Cond)...)
				offset := loopStart - (len(code) + 2)
				if offset > 255 {
					panic("TODO: for jump offset too big")
				}
				code = append(code, JumpTrue, Opcode(int8(offset)))
			}

			offset := len(code) - (forwardMark + 2)
			if offset > 255 {
				panic("TODO: for jump offset too big")
			}
			code[forwardMark+1] = Opcode(int8(offset))
		} else {
			offset := loopStart - (len(code) + 2)
			if offset > 255 {
				panic("TODO: for jump offset too big")
			}
			code = append(code, Jump, Opcode(int8(offset)))
		}

	//case *ast.ForInStmt:
	//
	//case *ast.ReturnStmt:
	//
	//case *ast.WhileStmt:
	//
	//case *ast.DoWhileStmt:
	//
	//case *ast.BreakStmt:
	//case *ast.ContinueStmt:
	//case *ast.NextStmt:
	//case *ast.ExitStmt:
	//
	//case *ast.DeleteStmt:
	//
	//case *ast.BlockStmt:

	default:
		// Should never happen
		panic(fmt.Sprintf("unexpected stmt type: %T", stmt))
	}
	return code
}

func (c *compiler) expr(expr ast.Expr) []Opcode {
	var code []Opcode
	switch e := expr.(type) {
	case *ast.NumExpr:
		if len(c.nums) > 255 {
			panic("TODO: too many nums!")
		}
		code = append(code, Num, Opcode(len(c.nums)))
		c.nums = append(c.nums, e.Value)

	case *ast.StrExpr:
		if len(c.strs) > 255 {
			panic("TODO: too many strs!")
		}
		code = append(code, Str, Opcode(len(c.strs)))
		c.strs = append(c.strs, e.Value)

	//case *ast.FieldExpr:
	//

	case *ast.VarExpr:
		if e.Index > 255 {
			panic("TODO: VarExpr index too big")
		}
		switch e.Scope {
		case ast.ScopeGlobal:
			code = append(code, Global, Opcode(e.Index))
		case ast.ScopeLocal:
		default: // ast.ScopeSpecial
		}

	//case *ast.RegExpr:
	//

	case *ast.BinaryExpr:
		switch e.Op {
		case lexer.AND:
			panic("TODO: &&")
		case lexer.OR:
			panic("TODO: ||")
		}
		code = append(code, c.expr(e.Left)...)
		code = append(code, c.expr(e.Right)...)
		var opcode Opcode
		switch e.Op {
		case lexer.ADD:
			opcode = Add
		case lexer.SUB:
			opcode = Sub
		case lexer.EQUALS:
			opcode = Equals
		case lexer.LESS:
			opcode = Less
		case lexer.LTE:
			opcode = LessOrEqual
		case lexer.CONCAT:
			opcode = Concat
		case lexer.MUL:
			opcode = Mul
		case lexer.DIV:
			opcode = Div
		case lexer.GREATER:
			opcode = Greater
		case lexer.GTE:
			opcode = GreaterOrEqual
		case lexer.NOT_EQUALS:
			opcode = NotEquals
		case lexer.MATCH:
			opcode = Match
		case lexer.NOT_MATCH:
			opcode = NotMatch
		case lexer.POW:
			opcode = Pow
		case lexer.MOD:
			opcode = Mod
		default:
			panic(fmt.Sprintf("unexpected binary operation: %s", e.Op))
		}
		code = append(code, opcode)

	//case *ast.IncrExpr:

	case *ast.AssignExpr:
		code = append(code, c.expr(e.Right)...)
		code = append(code, Dupe)
		switch left := e.Left.(type) {
		case *ast.VarExpr:
			if left.Index > 255 {
				panic("TODO: AssignExpr var index too big")
			}
			switch left.Scope {
			case ast.ScopeGlobal:
				code = append(code, AssignGlobal, Opcode(left.Index))
			case ast.ScopeLocal:
			default: // ast.ScopeSpecial
			}
		case *ast.IndexExpr:
		default: // *ast.FieldExpr
		}

	//case *ast.AugAssignExpr:
	//
	//case *ast.CondExpr:
	//
	//case *ast.IndexExpr:
	//
	//case *ast.CallExpr:
	//
	//case *ast.UnaryExpr:
	//
	//case *ast.InExpr:
	//
	//case *ast.UserCallExpr:
	//
	//case *ast.GetlineExpr:

	default:
		// Should never happen
		panic(fmt.Sprintf("unexpected expr type: %T", expr))
	}
	return code
}