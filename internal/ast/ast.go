// GoAWK parser - abstract syntax tree structs

package ast

import (
	"fmt"
	"strconv"
	"strings"

	. "github.com/nuvolaris/goawk/lexer"
)

// Program is an entire AWK program.
type Program struct {
	Begin     []Stmts
	Actions   []*Action
	End       []Stmts
	Functions []*Function
}

// ResolvedProgram is a parsed AWK program + additional data prepared by resolve step
// needed for subsequent interpretation
type ResolvedProgram struct {
	Program
	Scalars map[string]int
	Arrays  map[string]int
}

// String returns an indented, pretty-printed version of the parsed
// program.
func (p *Program) String() string {
	parts := []string{}
	for _, ss := range p.Begin {
		parts = append(parts, "BEGIN {\n"+ss.String()+"}")
	}
	for _, a := range p.Actions {
		parts = append(parts, a.String())
	}
	for _, ss := range p.End {
		parts = append(parts, "END {\n"+ss.String()+"}")
	}
	for _, function := range p.Functions {
		parts = append(parts, function.String())
	}
	return strings.Join(parts, "\n\n")
}

// Stmts is a block containing multiple statements.
type Stmts []Stmt

func (ss Stmts) String() string {
	lines := []string{}
	for _, s := range ss {
		subLines := strings.Split(s.String(), "\n")
		for _, sl := range subLines {
			lines = append(lines, "    "+sl+"\n")
		}
	}
	return strings.Join(lines, "")
}

// Action is pattern-action section of a program.
type Action struct {
	Pattern []Expr
	Stmts   Stmts
}

func (a *Action) String() string {
	patterns := make([]string, len(a.Pattern))
	for i, p := range a.Pattern {
		patterns[i] = p.String()
	}
	sep := ""
	if len(patterns) > 0 && a.Stmts != nil {
		sep = " "
	}
	stmtsStr := ""
	if a.Stmts != nil {
		stmtsStr = "{\n" + a.Stmts.String() + "}"
	}
	return strings.Join(patterns, ", ") + sep + stmtsStr
}

// Node is an interface to be satisfied by all AST elements.
// We need it to be able to work with AST in a generic way, like in ast.Walk().
type Node interface {
	node()
}

// All these types implement the Node interface.
func (p *Program) node()        {}
func (a *Action) node()         {}
func (f *Function) node()       {}
func (e *FieldExpr) node()      {}
func (e *NamedFieldExpr) node() {}
func (e *UnaryExpr) node()      {}
func (e *BinaryExpr) node()     {}
func (e *ArrayExpr) node()      {}
func (e *InExpr) node()         {}
func (e *CondExpr) node()       {}
func (e *NumExpr) node()        {}
func (e *StrExpr) node()        {}
func (e *RegExpr) node()        {}
func (e *VarExpr) node()        {}
func (e *IndexExpr) node()      {}
func (e *AssignExpr) node()     {}
func (e *AugAssignExpr) node()  {}
func (e *IncrExpr) node()       {}
func (e *CallExpr) node()       {}
func (e *UserCallExpr) node()   {}
func (e *MultiExpr) node()      {}
func (e *GetlineExpr) node()    {}
func (s *PrintStmt) node()      {}
func (s *PrintfStmt) node()     {}
func (s *ExprStmt) node()       {}
func (s *IfStmt) node()         {}
func (s *ForStmt) node()        {}
func (s *ForInStmt) node()      {}
func (s *WhileStmt) node()      {}
func (s *DoWhileStmt) node()    {}
func (s *BreakStmt) node()      {}
func (s *ContinueStmt) node()   {}
func (s *NextStmt) node()       {}
func (s *ExitStmt) node()       {}
func (s *DeleteStmt) node()     {}
func (s *ReturnStmt) node()     {}
func (s *BlockStmt) node()      {}

// Expr is the abstract syntax tree for any AWK expression.
type Expr interface {
	Node
	expr()
	String() string
}

// All these types implement the Expr interface.
func (e *FieldExpr) expr()      {}
func (e *NamedFieldExpr) expr() {}
func (e *UnaryExpr) expr()      {}
func (e *BinaryExpr) expr()     {}
func (e *ArrayExpr) expr()      {}
func (e *InExpr) expr()         {}
func (e *CondExpr) expr()       {}
func (e *NumExpr) expr()        {}
func (e *StrExpr) expr()        {}
func (e *RegExpr) expr()        {}
func (e *VarExpr) expr()        {}
func (e *IndexExpr) expr()      {}
func (e *AssignExpr) expr()     {}
func (e *AugAssignExpr) expr()  {}
func (e *IncrExpr) expr()       {}
func (e *CallExpr) expr()       {}
func (e *UserCallExpr) expr()   {}
func (e *MultiExpr) expr()      {}
func (e *GetlineExpr) expr()    {}

// FieldExpr is an expression like $0.
type FieldExpr struct {
	Index Expr
}

func (e *FieldExpr) String() string {
	return "$" + e.Index.String()
}

// NamedFieldExpr is an expression like @"name".
type NamedFieldExpr struct {
	Field Expr
}

func (e *NamedFieldExpr) String() string {
	return "@" + e.Field.String()
}

// UnaryExpr is an expression like -1234.
type UnaryExpr struct {
	Op    Token
	Value Expr
}

func (e *UnaryExpr) String() string {
	return e.Op.String() + e.Value.String()
}

// BinaryExpr is an expression like 1 + 2.
type BinaryExpr struct {
	Left  Expr
	Op    Token
	Right Expr
}

func (e *BinaryExpr) String() string {
	var opStr string
	if e.Op == CONCAT {
		opStr = " "
	} else {
		opStr = " " + e.Op.String() + " "
	}
	return "(" + e.Left.String() + opStr + e.Right.String() + ")"
}

// ArrayExpr is an array reference. Not really a stand-alone
// expression, except as an argument to split() or a user function
// call.
type ArrayExpr struct {
	Scope VarScope
	Index int
	Name  string
	Pos   Position
}

func (e *ArrayExpr) String() string {
	return e.Name
}

// InExpr is an expression like (index in array).
type InExpr struct {
	Index []Expr
	Array *ArrayExpr
}

func (e *InExpr) String() string {
	if len(e.Index) == 1 {
		return "(" + e.Index[0].String() + " in " + e.Array.String() + ")"
	}
	indices := make([]string, len(e.Index))
	for i, index := range e.Index {
		indices[i] = index.String()
	}
	return "((" + strings.Join(indices, ", ") + ") in " + e.Array.String() + ")"
}

// CondExpr is an expression like cond ? 1 : 0.
type CondExpr struct {
	Cond  Expr
	True  Expr
	False Expr
}

func (e *CondExpr) String() string {
	return "(" + e.Cond.String() + " ? " + e.True.String() + " : " + e.False.String() + ")"
}

// NumExpr is a literal number like 1234.
type NumExpr struct {
	Value float64
}

func (e *NumExpr) String() string {
	if e.Value == float64(int(e.Value)) {
		return strconv.Itoa(int(e.Value))
	} else {
		return fmt.Sprintf("%.6g", e.Value)
	}
}

// StrExpr is a literal string like "foo".
type StrExpr struct {
	Value string
}

func (e *StrExpr) String() string {
	return strconv.Quote(e.Value)
}

// RegExpr is a stand-alone regex expression, equivalent to:
// $0 ~ /regex/.
type RegExpr struct {
	Regex string
}

func (e *RegExpr) String() string {
	escaped := strings.Replace(e.Regex, "/", `\/`, -1)
	return "/" + escaped + "/"
}

// meaning it will be set during resolve step
const resolvedLater = -1

type VarScope int

const (
	ScopeSpecial VarScope = iota
	ScopeGlobal
	ScopeLocal
)

// VarExpr is a variable reference (special var, global, or local).
// Index is the resolved variable index used by the interpreter; Name
// is the original name used by String().
type VarExpr struct {
	Scope VarScope
	Index int
	Name  string
	Pos   Position
}

func (e *VarExpr) String() string {
	return e.Name
}

// IndexExpr is an expression like a[k] (rvalue or lvalue).
type IndexExpr struct {
	Array *ArrayExpr
	Index []Expr
}

func (e *IndexExpr) String() string {
	indices := make([]string, len(e.Index))
	for i, index := range e.Index {
		indices[i] = index.String()
	}
	return e.Array.String() + "[" + strings.Join(indices, ", ") + "]"
}

// AssignExpr is an expression like x = 1234.
type AssignExpr struct {
	Left  Expr // can be one of: var, array[x], $n
	Right Expr
}

func (e *AssignExpr) String() string {
	return e.Left.String() + " = " + e.Right.String()
}

// AugAssignExpr is an assignment expression like x += 5.
type AugAssignExpr struct {
	Left  Expr // can be one of: var, array[x], $n
	Op    Token
	Right Expr
}

func (e *AugAssignExpr) String() string {
	return e.Left.String() + " " + e.Op.String() + "= " + e.Right.String()
}

// IncrExpr is an increment or decrement expression like x++ or --y.
type IncrExpr struct {
	Expr Expr
	Op   Token
	Pre  bool
}

func (e *IncrExpr) String() string {
	if e.Pre {
		return e.Op.String() + e.Expr.String()
	} else {
		return e.Expr.String() + e.Op.String()
	}
}

// CallExpr is a builtin function call like length($1).
type CallExpr struct {
	Func Token
	Args []Expr
}

func (e *CallExpr) String() string {
	args := make([]string, len(e.Args))
	for i, a := range e.Args {
		args[i] = a.String()
	}
	return e.Func.String() + "(" + strings.Join(args, ", ") + ")"
}

// UserCallExpr is a user-defined function call like my_func(1, 2, 3)
//
// Index is the resolved function index used by the interpreter; Name
// is the original name used by String().
type UserCallExpr struct {
	Native bool // false = AWK-defined function, true = native Go func
	Index  int
	Name   string
	Args   []Expr
	Pos    Position
}

func (e *UserCallExpr) String() string {
	args := make([]string, len(e.Args))
	for i, a := range e.Args {
		args[i] = a.String()
	}
	return e.Name + "(" + strings.Join(args, ", ") + ")"
}

// MultiExpr isn't an interpretable expression, but it's used as a
// pseudo-expression for print[f] parsing.
type MultiExpr struct {
	Exprs []Expr
}

func (e *MultiExpr) String() string {
	exprs := make([]string, len(e.Exprs))
	for i, e := range e.Exprs {
		exprs[i] = e.String()
	}
	return "(" + strings.Join(exprs, ", ") + ")"
}

// GetlineExpr is an expression read from file or pipe input.
type GetlineExpr struct {
	Command Expr
	Target  Expr
	File    Expr
}

func (e *GetlineExpr) String() string {
	s := ""
	if e.Command != nil {
		s += e.Command.String() + " |"
	}
	s += "getline"
	if e.Target != nil {
		s += " " + e.Target.String()
	}
	if e.File != nil {
		s += " <" + e.File.String()
	}
	return s
}

// IsLValue returns true if the given expression can be used as an
// lvalue (on the left-hand side of an assignment, in a ++ or --
// operation, or as the third argument to sub or gsub).
func IsLValue(expr Expr) bool {
	switch expr.(type) {
	case *VarExpr, *IndexExpr, *FieldExpr:
		return true
	default:
		return false
	}
}

// Stmt is the abstract syntax tree for any AWK statement.
type Stmt interface {
	Node
	StartPos() Position // position of first character belonging to the node
	EndPos() Position   // position of first character immediately after the node
	stmt()
	String() string
}

// All these types implement the Stmt interface.
func (s *PrintStmt) stmt()    {}
func (s *PrintfStmt) stmt()   {}
func (s *ExprStmt) stmt()     {}
func (s *IfStmt) stmt()       {}
func (s *ForStmt) stmt()      {}
func (s *ForInStmt) stmt()    {}
func (s *WhileStmt) stmt()    {}
func (s *DoWhileStmt) stmt()  {}
func (s *BreakStmt) stmt()    {}
func (s *ContinueStmt) stmt() {}
func (s *NextStmt) stmt()     {}
func (s *ExitStmt) stmt()     {}
func (s *DeleteStmt) stmt()   {}
func (s *ReturnStmt) stmt()   {}
func (s *BlockStmt) stmt()    {}

func (s *PrintStmt) StartPos() Position    { return s.Start }
func (s *PrintfStmt) StartPos() Position   { return s.Start }
func (s *ExprStmt) StartPos() Position     { return s.Start }
func (s *IfStmt) StartPos() Position       { return s.Start }
func (s *ForStmt) StartPos() Position      { return s.Start }
func (s *ForInStmt) StartPos() Position    { return s.Start }
func (s *WhileStmt) StartPos() Position    { return s.Start }
func (s *DoWhileStmt) StartPos() Position  { return s.Start }
func (s *BreakStmt) StartPos() Position    { return s.Start }
func (s *ContinueStmt) StartPos() Position { return s.Start }
func (s *NextStmt) StartPos() Position     { return s.Start }
func (s *ExitStmt) StartPos() Position     { return s.Start }
func (s *DeleteStmt) StartPos() Position   { return s.Start }
func (s *ReturnStmt) StartPos() Position   { return s.Start }
func (s *BlockStmt) StartPos() Position    { return s.Start }

func (s *PrintStmt) EndPos() Position    { return s.End }
func (s *PrintfStmt) EndPos() Position   { return s.End }
func (s *ExprStmt) EndPos() Position     { return s.End }
func (s *IfStmt) EndPos() Position       { return s.End }
func (s *ForStmt) EndPos() Position      { return s.End }
func (s *ForInStmt) EndPos() Position    { return s.End }
func (s *WhileStmt) EndPos() Position    { return s.End }
func (s *DoWhileStmt) EndPos() Position  { return s.End }
func (s *BreakStmt) EndPos() Position    { return s.End }
func (s *ContinueStmt) EndPos() Position { return s.End }
func (s *NextStmt) EndPos() Position     { return s.End }
func (s *ExitStmt) EndPos() Position     { return s.End }
func (s *DeleteStmt) EndPos() Position   { return s.End }
func (s *ReturnStmt) EndPos() Position   { return s.End }
func (s *BlockStmt) EndPos() Position    { return s.End }

// PrintStmt is a statement like print $1, $3.
type PrintStmt struct {
	Args     []Expr
	Redirect Token
	Dest     Expr
	Start    Position
	End      Position
}

func (s *PrintStmt) String() string {
	return printString("print", s.Args, s.Redirect, s.Dest)
}

func printString(f string, args []Expr, redirect Token, dest Expr) string {
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = a.String()
	}
	str := f + " " + strings.Join(parts, ", ")
	if dest != nil {
		str += " " + redirect.String() + dest.String()
	}
	return str
}

// PrintfStmt is a statement like printf "%3d", 1234.
type PrintfStmt struct {
	Args     []Expr
	Redirect Token
	Dest     Expr
	Start    Position
	End      Position
}

func (s *PrintfStmt) String() string {
	return printString("printf", s.Args, s.Redirect, s.Dest)
}

// ExprStmt is statement like a bare function call: my_func(x).
type ExprStmt struct {
	Expr  Expr
	Start Position
	End   Position
}

func (s *ExprStmt) String() string {
	return s.Expr.String()
}

// IfStmt is an if or if-else statement.
type IfStmt struct {
	Cond      Expr
	BodyStart Position
	Body      Stmts
	Else      Stmts
	Start     Position
	End       Position
}

func (s *IfStmt) String() string {
	str := "if (" + trimParens(s.Cond.String()) + ") {\n" + s.Body.String() + "}"
	if len(s.Else) > 0 {
		str += " else {\n" + s.Else.String() + "}"
	}
	return str
}

// ForStmt is a C-like for loop: for (i=0; i<10; i++) print i.
type ForStmt struct {
	Pre       Stmt
	Cond      Expr
	Post      Stmt
	BodyStart Position
	Body      Stmts
	Start     Position
	End       Position
}

func (s *ForStmt) String() string {
	preStr := ""
	if s.Pre != nil {
		preStr = s.Pre.String()
	}
	condStr := ""
	if s.Cond != nil {
		condStr = " " + trimParens(s.Cond.String())
	}
	postStr := ""
	if s.Post != nil {
		postStr = " " + s.Post.String()
	}
	return "for (" + preStr + ";" + condStr + ";" + postStr + ") {\n" + s.Body.String() + "}"
}

// ForInStmt is a for loop like for (k in a) print k, a[k].
type ForInStmt struct {
	Var       *VarExpr
	Array     *ArrayExpr
	BodyStart Position
	Body      Stmts
	Start     Position
	End       Position
}

func (s *ForInStmt) String() string {
	return "for (" + s.Var.String() + " in " + s.Array.String() + ") {\n" + s.Body.String() + "}"
}

// WhileStmt is a while loop.
type WhileStmt struct {
	Cond      Expr
	BodyStart Position
	Body      Stmts
	Start     Position
	End       Position
}

func (s *WhileStmt) String() string {
	return "while (" + trimParens(s.Cond.String()) + ") {\n" + s.Body.String() + "}"
}

// DoWhileStmt is a do-while loop.
type DoWhileStmt struct {
	Body  Stmts
	Cond  Expr
	Start Position
	End   Position
}

func (s *DoWhileStmt) String() string {
	return "do {\n" + s.Body.String() + "} while (" + trimParens(s.Cond.String()) + ")"
}

// BreakStmt is a break statement.
type BreakStmt struct {
	Start Position
	End   Position
}

func (s *BreakStmt) String() string {
	return "break"
}

// ContinueStmt is a continue statement.
type ContinueStmt struct {
	Start Position
	End   Position
}

func (s *ContinueStmt) String() string {
	return "continue"
}

// NextStmt is a next statement.
type NextStmt struct {
	Start Position
	End   Position
}

func (s *NextStmt) String() string {
	return "next"
}

// ExitStmt is an exit statement.
type ExitStmt struct {
	Status Expr
	Start  Position
	End    Position
}

func (s *ExitStmt) String() string {
	var statusStr string
	if s.Status != nil {
		statusStr = " " + s.Status.String()
	}
	return "exit" + statusStr
}

// DeleteStmt is a statement like delete a[k].
type DeleteStmt struct {
	Array *ArrayExpr
	Index []Expr
	Start Position
	End   Position
}

func (s *DeleteStmt) String() string {
	indices := make([]string, len(s.Index))
	for i, index := range s.Index {
		indices[i] = index.String()
	}
	return "delete " + s.Array.String() + "[" + strings.Join(indices, ", ") + "]"
}

// ReturnStmt is a return statement.
type ReturnStmt struct {
	Value Expr
	Start Position
	End   Position
}

func (s *ReturnStmt) String() string {
	var valueStr string
	if s.Value != nil {
		valueStr = " " + s.Value.String()
	}
	return "return" + valueStr
}

// BlockStmt is a stand-alone block like { print "x" }.
type BlockStmt struct {
	Body  Stmts
	Start Position
	End   Position
}

func (s *BlockStmt) String() string {
	return "{\n" + s.Body.String() + "}"
}

// Function is the AST for a user-defined function.
type Function struct {
	Name   string
	Params []string
	Arrays []bool
	Body   Stmts
	Pos    Position
}

func (f *Function) String() string {
	return "function " + f.Name + "(" + strings.Join(f.Params, ", ") + ") {\n" +
		f.Body.String() + "}"
}

func trimParens(s string) string {
	if strings.HasPrefix(s, "(") && strings.HasSuffix(s, ")") {
		s = s[1 : len(s)-1]
	}
	return s
}

// VarRef is a constructor for *VarExpr
func VarRef(name string, pos Position) *VarExpr {
	return &VarExpr{resolvedLater, resolvedLater, name, pos}
}

// ArrayRef is a constructor for *ArrayExpr
func ArrayRef(name string, pos Position) *ArrayExpr {
	return &ArrayExpr{resolvedLater, resolvedLater, name, pos}
}

// UserCall is a constructor for *UserCallExpr
func UserCall(name string, args []Expr, pos Position) *UserCallExpr {
	return &UserCallExpr{false, resolvedLater, name, args, pos}
}

// PositionError represents an error bound to specific position in source.
type PositionError struct {
	// Source line/column position where the error occurred.
	Position Position
	// Error message.
	Message string
}

// PosErrorf like fmt.Errorf, but with an explicit position.
func PosErrorf(pos Position, format string, args ...interface{}) error {
	message := fmt.Sprintf(format, args...)
	return &PositionError{pos, message}
}

// Error returns a formatted version of the error, including the line
// and column numbers.
func (e *PositionError) Error() string {
	return fmt.Sprintf("parse error at %d:%d: %s", e.Position.Line, e.Position.Column, e.Message)
}
