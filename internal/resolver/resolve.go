// Resolve function calls and variable types
package resolver

import (
	"fmt"
	"io"
	"reflect"
	"sort"

	"github.com/nuvolaris/goawk/internal/ast"
	. "github.com/nuvolaris/goawk/lexer"
)

type resolver struct {
	// Resolving state
	funcName string // function name if parsing a func, else ""

	// Variable tracking and resolving
	locals    map[string]bool                // current function's locals (for determining scope)
	varTypes  map[string]map[string]typeInfo // map of func name to var name to type
	varRefs   []varRef                       // all variable references (usually scalars)
	arrayRefs []arrayRef                     // all array references

	// Function tracking
	functions   map[string]int // map of function name to index
	userCalls   []userCall     // record calls so we can resolve them later
	nativeFuncs map[string]interface{}

	// Configuration and debugging
	debugTypes  bool      // show variable types for debugging
	debugWriter io.Writer // where the debug output goes
}

type Config struct {
	// Enable printing of type information
	DebugTypes bool

	// io.Writer to print type information on (for example, os.Stderr)
	DebugWriter io.Writer

	// Map of named Go functions to allow calling from AWK. See docs
	// on interp.Config.Funcs for details.
	Funcs map[string]interface{}
}

func Resolve(prog *ast.Program, config *Config) *ast.ResolvedProgram {
	r := newResolver(config)

	resolvedProg := &ast.ResolvedProgram{Program: *prog}

	ast.Walk(r, prog)

	r.resolveUserCalls(prog)
	r.resolveVars(resolvedProg)

	return resolvedProg
}

func (r *resolver) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {

	case *ast.Function:
		function := n
		name := function.Name
		if _, ok := r.functions[name]; ok {
			panic(ast.PosErrorf(function.Pos, "function %q already defined", name))
		}
		r.functions[name] = len(r.functions)
		r.locals = make(map[string]bool, 7)
		for _, param := range function.Params {
			r.locals[param] = true
		}
		r.funcName = name
		r.varTypes[name] = make(map[string]typeInfo)

		ast.WalkStmtList(r, function.Body)

		r.funcName = ""
		r.locals = nil

	case *ast.VarExpr:
		r.recordVarRef(n)

	case *ast.ArrayExpr:
		r.recordArrayRef(n)

	case *ast.UserCallExpr:
		name := n.Name
		if r.locals[name] {
			panic(ast.PosErrorf(n.Pos, "can't call local variable %q as function", name))
		}
		for i, arg := range n.Args {
			ast.Walk(r, arg)
			r.processUserCallArg(name, arg, i)
		}
		r.userCalls = append(r.userCalls, userCall{n, n.Pos, r.funcName})
	default:
		return r
	}

	return nil
}

type varType int

const (
	typeUnknown varType = iota
	typeScalar
	typeArray
)

func (t varType) String() string {
	switch t {
	case typeScalar:
		return "Scalar"
	case typeArray:
		return "Array"
	default:
		return "Unknown"
	}
}

// typeInfo records type information for a single variable
type typeInfo struct {
	typ      varType
	ref      *ast.VarExpr
	scope    ast.VarScope
	index    int
	callName string
	argIndex int
}

// Used by printVarTypes when debugTypes is turned on
func (t typeInfo) String() string {
	var scope string
	switch t.scope {
	case ast.ScopeGlobal:
		scope = "Global"
	case ast.ScopeLocal:
		scope = "Local"
	default:
		scope = "Special"
	}
	return fmt.Sprintf("typ=%s ref=%p scope=%s index=%d callName=%q argIndex=%d",
		t.typ, t.ref, scope, t.index, t.callName, t.argIndex)
}

// A single variable reference (normally scalar)
type varRef struct {
	funcName string
	ref      *ast.VarExpr
	isArg    bool
}

// A single array reference
type arrayRef struct {
	funcName string
	ref      *ast.ArrayExpr
}

// Constructs the resolver
func newResolver(config *Config) *resolver {
	r := &resolver{}
	if config != nil {
		r.nativeFuncs = config.Funcs
		r.debugTypes = config.DebugTypes
		r.debugWriter = config.DebugWriter
	}
	r.varTypes = make(map[string]map[string]typeInfo)
	r.varTypes[""] = make(map[string]typeInfo) // globals
	r.functions = make(map[string]int)
	initialPos := Position{1, 1}
	r.recordArrayRef(ast.ArrayRef("ARGV", initialPos))    // interpreter relies on ARGV being present
	r.recordArrayRef(ast.ArrayRef("ENVIRON", initialPos)) // and other built-in arrays
	r.recordArrayRef(ast.ArrayRef("FIELDS", initialPos))
	return r
}

// Records a call to a user function (for resolving indexes later)
type userCall struct {
	call   *ast.UserCallExpr
	pos    Position
	inFunc string
}

// After parsing, resolve all user calls to their indexes. Also
// ensures functions called have actually been defined, and that
// they're not being called with too many arguments.
func (r *resolver) resolveUserCalls(prog *ast.Program) {
	// Number the native funcs (order by name to get consistent order)
	nativeNames := make([]string, 0, len(r.nativeFuncs))
	for name := range r.nativeFuncs {
		nativeNames = append(nativeNames, name)
	}
	sort.Strings(nativeNames)
	nativeIndexes := make(map[string]int, len(nativeNames))
	for i, name := range nativeNames {
		nativeIndexes[name] = i
	}

	for _, c := range r.userCalls {
		// AWK-defined functions take precedence over native Go funcs
		index, ok := r.functions[c.call.Name]
		if !ok {
			f, haveNative := r.nativeFuncs[c.call.Name]
			if !haveNative {
				panic(ast.PosErrorf(c.pos, "undefined function %q", c.call.Name))
			}
			typ := reflect.TypeOf(f)
			if !typ.IsVariadic() && len(c.call.Args) > typ.NumIn() {
				panic(ast.PosErrorf(c.pos, "%q called with more arguments than declared", c.call.Name))
			}
			c.call.Native = true
			c.call.Index = nativeIndexes[c.call.Name]
			continue
		}
		function := prog.Functions[index]
		if len(c.call.Args) > len(function.Params) {
			panic(ast.PosErrorf(c.pos, "%q called with more arguments than declared", c.call.Name))
		}
		c.call.Index = index
	}
}

// For arguments that are variable references, we don't know the
// type based on context, so mark the types for these as unknown.
func (r *resolver) processUserCallArg(funcName string, arg ast.Expr, index int) {
	if varExpr, ok := arg.(*ast.VarExpr); ok {
		scope, varFuncName := r.getScope(varExpr.Name)
		ref := r.varTypes[varFuncName][varExpr.Name].ref
		if ref == varExpr {
			// Only applies if this is the first reference to this
			// variable (otherwise we know the type already)
			r.varTypes[varFuncName][varExpr.Name] = typeInfo{typeUnknown, ref, scope, 0, funcName, index}
		}
		// Mark the last related varRef (the most recent one) as a
		// call argument for later error handling
		r.varRefs[len(r.varRefs)-1].isArg = true
	}
}

// Determine scope of given variable reference (and funcName if it's
// a local, otherwise empty string)
func (r *resolver) getScope(name string) (ast.VarScope, string) {
	switch {
	case r.locals[name]:
		return ast.ScopeLocal, r.funcName
	case ast.SpecialVarIndex(name) > 0:
		return ast.ScopeSpecial, ""
	default:
		return ast.ScopeGlobal, ""
	}
}

// Record a variable (scalar) reference and return the *VarExpr (but
// VarExpr.Index won't be set till later)
func (r *resolver) recordVarRef(expr *ast.VarExpr) {
	name := expr.Name
	scope, funcName := r.getScope(name)
	expr.Scope = scope
	r.varRefs = append(r.varRefs, varRef{funcName, expr, false})
	info := r.varTypes[funcName][name]
	if info.typ == typeUnknown {
		r.varTypes[funcName][name] = typeInfo{typeScalar, expr, scope, 0, info.callName, 0}
	}
}

// Record an array reference and return the *ArrayExpr (but
// ArrayExpr.Index won't be set till later)
func (r *resolver) recordArrayRef(expr *ast.ArrayExpr) {
	name := expr.Name
	scope, funcName := r.getScope(name)
	if scope == ast.ScopeSpecial {
		panic(ast.PosErrorf(expr.Pos, "can't use scalar %q as array", name))
	}
	expr.Scope = scope
	r.arrayRefs = append(r.arrayRefs, arrayRef{funcName, expr})
	info := r.varTypes[funcName][name]
	if info.typ == typeUnknown {
		r.varTypes[funcName][name] = typeInfo{typeArray, nil, scope, 0, info.callName, 0}
	}
}

// Print variable type information (for debugging) on p.debugWriter
func (r *resolver) printVarTypes(prog *ast.ResolvedProgram) {
	fmt.Fprintf(r.debugWriter, "scalars: %v\n", prog.Scalars)
	fmt.Fprintf(r.debugWriter, "arrays: %v\n", prog.Arrays)
	funcNames := []string{}
	for funcName := range r.varTypes {
		funcNames = append(funcNames, funcName)
	}
	sort.Strings(funcNames)
	for _, funcName := range funcNames {
		if funcName != "" {
			fmt.Fprintf(r.debugWriter, "function %s\n", funcName)
		} else {
			fmt.Fprintf(r.debugWriter, "globals\n")
		}
		varNames := []string{}
		for name := range r.varTypes[funcName] {
			varNames = append(varNames, name)
		}
		sort.Strings(varNames)
		for _, name := range varNames {
			info := r.varTypes[funcName][name]
			fmt.Fprintf(r.debugWriter, "  %s: %s\n", name, info)
		}
	}
}

// Resolve unknown variables types and generate variable indexes and
// name-to-index mappings for interpreter
func (r *resolver) resolveVars(prog *ast.ResolvedProgram) {
	// First go through all unknown types and try to determine the
	// type from the parameter type in that function definition.
	// Iterate through functions in topological order, for example
	// if f() calls g(), process g first, then f.
	callGraph := make(map[string]map[string]struct{})
	for _, call := range r.userCalls {
		if _, ok := callGraph[call.inFunc]; !ok {
			callGraph[call.inFunc] = make(map[string]struct{})
		}
		callGraph[call.inFunc][call.call.Name] = struct{}{}
	}
	sortedFuncs := topoSort(callGraph)
	for _, funcName := range sortedFuncs {
		infos := r.varTypes[funcName]
		for name, info := range infos {
			if info.scope == ast.ScopeSpecial || info.typ != typeUnknown {
				// It's a special var or type is already known
				continue
			}
			funcIndex, ok := r.functions[info.callName]
			if !ok {
				// Function being called is a native function
				continue
			}
			// Determine var type based on type of this parameter
			// in the called function (if we know that)
			paramName := prog.Functions[funcIndex].Params[info.argIndex]
			typ := r.varTypes[info.callName][paramName].typ
			if typ != typeUnknown {
				if r.debugTypes {
					fmt.Fprintf(r.debugWriter, "resolving %s:%s to %s\n",
						funcName, name, typ)
				}
				info.typ = typ
				r.varTypes[funcName][name] = info
			}
		}
	}

	// Resolve global variables (iteration order is undefined, so
	// assign indexes basically randomly)
	prog.Scalars = make(map[string]int)
	prog.Arrays = make(map[string]int)
	for name, info := range r.varTypes[""] {
		_, isFunc := r.functions[name]
		if isFunc {
			// Global var can't also be the name of a function
			pos := Position{1, 1} // Ideally it'd be the position of the global var.
			panic(ast.PosErrorf(pos, "global var %q can't also be a function", name))
		}
		var index int
		if info.scope == ast.ScopeSpecial {
			index = ast.SpecialVarIndex(name)
		} else if info.typ == typeArray {
			index = len(prog.Arrays)
			prog.Arrays[name] = index
		} else {
			index = len(prog.Scalars)
			prog.Scalars[name] = index
		}
		info.index = index
		r.varTypes[""][name] = info
	}

	// Fill in unknown parameter types that are being called with arrays,
	// for example, as in the following code:
	//
	// BEGIN { arr[0]; f(arr) }
	// function f(a) { }
	for _, c := range r.userCalls {
		if c.call.Native {
			continue
		}
		function := prog.Functions[c.call.Index]
		for i, arg := range c.call.Args {
			varExpr, ok := arg.(*ast.VarExpr)
			if !ok {
				continue
			}
			funcName := r.getVarFuncName(prog, varExpr.Name, c.inFunc)
			argType := r.varTypes[funcName][varExpr.Name]
			paramType := r.varTypes[function.Name][function.Params[i]]
			if argType.typ == typeArray && paramType.typ == typeUnknown {
				paramType.typ = argType.typ
				r.varTypes[function.Name][function.Params[i]] = paramType
			}
		}
	}

	// Resolve local variables (assign indexes in order of params).
	// Also patch up Function.Arrays (tells interpreter which args
	// are arrays).
	for funcName, infos := range r.varTypes {
		if funcName == "" {
			continue
		}
		scalarIndex := 0
		arrayIndex := 0
		functionIndex := r.functions[funcName]
		function := prog.Functions[functionIndex]
		arrays := make([]bool, len(function.Params))
		for i, name := range function.Params {
			info := infos[name]
			var index int
			if info.typ == typeArray {
				index = arrayIndex
				arrayIndex++
				arrays[i] = true
			} else {
				// typeScalar or typeUnknown: variables may still be
				// of unknown type if they've never been referenced --
				// default to scalar in that case
				index = scalarIndex
				scalarIndex++
			}
			info.index = index
			r.varTypes[funcName][name] = info
		}
		prog.Functions[functionIndex].Arrays = arrays
	}

	// Check that variables passed to functions are the correct type
	for _, c := range r.userCalls {
		// Check native function calls
		if c.call.Native {
			for _, arg := range c.call.Args {
				varExpr, ok := arg.(*ast.VarExpr)
				if !ok {
					// Non-variable expression, must be scalar
					continue
				}
				funcName := r.getVarFuncName(prog, varExpr.Name, c.inFunc)
				info := r.varTypes[funcName][varExpr.Name]
				if info.typ == typeArray {
					panic(ast.PosErrorf(c.pos, "can't pass array %q to native function", varExpr.Name))
				}
			}
			continue
		}

		// Check AWK function calls
		function := prog.Functions[c.call.Index]
		for i, arg := range c.call.Args {
			varExpr, ok := arg.(*ast.VarExpr)
			if !ok {
				if function.Arrays[i] {
					panic(ast.PosErrorf(c.pos, "can't pass scalar %s as array param", arg))
				}
				continue
			}
			funcName := r.getVarFuncName(prog, varExpr.Name, c.inFunc)
			info := r.varTypes[funcName][varExpr.Name]
			if info.typ == typeArray && !function.Arrays[i] {
				panic(ast.PosErrorf(c.pos, "can't pass array %q as scalar param", varExpr.Name))
			}
			if info.typ != typeArray && function.Arrays[i] {
				panic(ast.PosErrorf(c.pos, "can't pass scalar %q as array param", varExpr.Name))
			}
		}
	}

	if r.debugTypes {
		r.printVarTypes(prog)
	}

	// Patch up variable indexes (interpreter uses an index instead
	// of name for more efficient lookups)
	for _, varRef := range r.varRefs {
		info := r.varTypes[varRef.funcName][varRef.ref.Name]
		if info.typ == typeArray && !varRef.isArg {
			panic(ast.PosErrorf(varRef.ref.Pos, "can't use array %q as scalar", varRef.ref.Name))
		}
		varRef.ref.Index = info.index
	}
	for _, arrayRef := range r.arrayRefs {
		info := r.varTypes[arrayRef.funcName][arrayRef.ref.Name]
		if info.typ == typeScalar {
			panic(ast.PosErrorf(arrayRef.ref.Pos, "can't use scalar %q as array", arrayRef.ref.Name))
		}
		arrayRef.ref.Index = info.index
	}
}

// If name refers to a local (in function inFunc), return that
// function's name, otherwise return "" (meaning global).
func (r *resolver) getVarFuncName(prog *ast.ResolvedProgram, name, inFunc string) string {
	if inFunc == "" {
		return ""
	}
	for _, param := range prog.Functions[r.functions[inFunc]].Params {
		if name == param {
			return inFunc
		}
	}
	return ""
}
