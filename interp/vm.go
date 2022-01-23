package interp

import (
	"io"
	"math"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/benhoyt/goawk/internal/ast"
	"github.com/benhoyt/goawk/internal/compiler"
	"github.com/benhoyt/goawk/lexer"
)

// Execute a block of virtual machine instructions. This is a simple loop
// that, for every instruction, looks up the function in an array indexed by
// opcode and calls it.
func (p *interp) execute(code []compiler.Opcode) error {
	for i := 0; i < len(code); {
		op := code[i]
		i++

		n, err := vmFuncs[op](p, code, i)
		if err != nil {
			return err
		}
		i += n
	}
	return nil
}

// Type of function called for each instruction. Each function returns the
// number of arguments the instruction read from code[i:].
type vmFunc func(p *interp, code []compiler.Opcode, i int) (int, error)

var vmFuncs [compiler.EndOpcode]vmFunc

func init() {
	vmFuncs = [compiler.EndOpcode]vmFunc{
		compiler.Nop:                  vmNop,
		compiler.Num:                  vmNum,
		compiler.Str:                  vmStr,
		compiler.Dupe:                 vmDupe,
		compiler.Drop:                 vmDrop,
		compiler.Swap:                 vmSwap,
		compiler.Field:                vmField,
		compiler.FieldNum:             vmFieldNum,
		compiler.Global:               vmGlobal,
		compiler.Local:                vmLocal,
		compiler.Special:              vmSpecial,
		compiler.ArrayGlobal:          vmArrayGlobal,
		compiler.ArrayLocal:           vmArrayLocal,
		compiler.InGlobal:             vmInGlobal,
		compiler.InLocal:              vmInLocal,
		compiler.AssignField:          vmAssignField,
		compiler.AssignGlobal:         vmAssignGlobal,
		compiler.AssignLocal:          vmAssignLocal,
		compiler.AssignSpecial:        vmAssignSpecial,
		compiler.AssignArrayGlobal:    vmAssignArrayGlobal,
		compiler.AssignArrayLocal:     vmAssignArrayLocal,
		compiler.Delete:               vmDelete,
		compiler.DeleteAll:            vmDeleteAll,
		compiler.IncrField:            vmIncrField,
		compiler.IncrGlobal:           vmIncrGlobal,
		compiler.IncrLocal:            vmIncrLocal,
		compiler.IncrSpecial:          vmIncrSpecial,
		compiler.IncrArrayGlobal:      vmIncrArrayGlobal,
		compiler.IncrArrayLocal:       vmIncrArrayLocal,
		compiler.AugAssignField:       vmAugAssignField,
		compiler.AugAssignGlobal:      vmAugAssignGlobal,
		compiler.AugAssignLocal:       vmAugAssignLocal,
		compiler.AugAssignSpecial:     vmAugAssignSpecial,
		compiler.AugAssignArrayGlobal: vmAugAssignArrayGlobal,
		compiler.AugAssignArrayLocal:  vmAugAssignArrayLocal,
		compiler.Regex:                vmRegex,
		compiler.MultiIndex:           vmMultiIndex,
		compiler.Add:                  vmAdd,
		compiler.Subtract:             vmSubtract,
		compiler.Multiply:             vmMultiply,
		compiler.Divide:               vmDivide,
		compiler.Power:                vmPower,
		compiler.Modulo:               vmModulo,
		compiler.Equals:               vmEquals,
		compiler.NotEquals:            vmNotEquals,
		compiler.Less:                 vmLess,
		compiler.Greater:              vmGreater,
		compiler.LessOrEqual:          vmLessOrEqual,
		compiler.GreaterOrEqual:       vmGreaterOrEqual,
		compiler.Concat:               vmConcat,
		compiler.Match:                vmMatch,
		compiler.NotMatch:             vmNotMatch,
		compiler.Not:                  vmNot,
		compiler.UnaryMinus:           vmUnaryMinus,
		compiler.UnaryPlus:            vmUnaryPlus,
		compiler.Boolean:              vmBoolean,
		compiler.Jump:                 vmJump,
		compiler.JumpFalse:            vmJumpFalse,
		compiler.JumpTrue:             vmJumpTrue,
		compiler.JumpEquals:           vmJumpEquals,
		compiler.JumpNotEquals:        vmJumpNotEquals,
		compiler.JumpLess:             vmJumpLess,
		compiler.JumpGreater:          vmJumpGreater,
		compiler.JumpLessOrEqual:      vmJumpLessOrEqual,
		compiler.JumpGreaterOrEqual:   vmJumpGreaterOrEqual,
		compiler.Next:                 vmNext,
		compiler.Exit:                 vmExit,
		compiler.ForIn:                vmForIn,
		compiler.BreakForIn:           vmBreakForIn,
		compiler.CallAtan2:            vmCallAtan2,
		compiler.CallClose:            vmCallClose,
		compiler.CallCos:              vmCallCos,
		compiler.CallExp:              vmCallExp,
		compiler.CallFflush:           vmCallFflush,
		compiler.CallFflushAll:        vmCallFflushAll,
		compiler.CallGsub:             vmCallGsub,
		compiler.CallIndex:            vmCallIndex,
		compiler.CallInt:              vmCallInt,
		compiler.CallLength:           vmCallLength,
		compiler.CallLengthArg:        vmCallLengthArg,
		compiler.CallLog:              vmCallLog,
		compiler.CallMatch:            vmCallMatch,
		compiler.CallRand:             vmCallRand,
		compiler.CallSin:              vmCallSin,
		compiler.CallSplit:            vmCallSplit,
		compiler.CallSplitSep:         vmCallSplitSep,
		compiler.CallSprintf:          vmCallSprintf,
		compiler.CallSqrt:             vmCallSqrt,
		compiler.CallSrand:            vmCallSrand,
		compiler.CallSrandSeed:        vmCallSrandSeed,
		compiler.CallSub:              vmCallSub,
		compiler.CallSubstr:           vmCallSubstr,
		compiler.CallSubstrLength:     vmCallSubstrLength,
		compiler.CallSystem:           vmCallSystem,
		compiler.CallTolower:          vmCallTolower,
		compiler.CallToupper:          vmCallToupper,
		compiler.CallUser:             vmCallUser,
		compiler.CallNative:           vmCallNative,
		compiler.Return:               vmReturn,
		compiler.ReturnNull:           vmReturnNull,
		compiler.Nulls:                vmNulls,
		compiler.Print:                vmPrint,
		compiler.Printf:               vmPrintf,
		compiler.Getline:              vmGetline,
		compiler.GetlineField:         vmGetlineField,
		compiler.GetlineGlobal:        vmGetlineGlobal,
		compiler.GetlineLocal:         vmGetlineLocal,
		compiler.GetlineSpecial:       vmGetlineSpecial,
		compiler.GetlineArray:         vmGetlineArray,
	}
}

// Not used, but here for completeness.
func vmNop(p *interp, code []compiler.Opcode, i int) (int, error) {
	return 0, nil
}

func vmNum(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := code[i]
	p.push(num(p.nums[index]))
	return 1, nil
}

func vmStr(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := code[i]
	p.push(str(p.strs[index]))
	return 1, nil
}

func vmDupe(p *interp, code []compiler.Opcode, i int) (int, error) {
	v := p.peekTop()
	p.push(v)
	return 0, nil
}

func vmDrop(p *interp, code []compiler.Opcode, i int) (int, error) {
	_ = p.pop()
	return 0, nil
}

func vmSwap(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekTwo()
	p.replaceTwo(r, l)
	return 0, nil
}

func vmField(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := p.peekTop()
	v, err := p.getField(int(index.num()))
	if err != nil {
		return 0, err
	}
	p.replaceTop(v)
	return 0, nil
}

func vmFieldNum(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := code[i]
	v, err := p.getField(int(index))
	if err != nil {
		return 0, err
	}
	p.push(v)
	return 1, nil
}

func vmGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := code[i]
	p.push(p.globals[index])
	return 1, nil
}

func vmLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := code[i]
	p.push(p.frame[index])
	return 1, nil
}

func vmSpecial(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := code[i]
	p.push(p.getSpecial(int(index)))
	return 1, nil
}

func vmArrayGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayIndex := code[i]
	array := p.arrays[arrayIndex]
	index := p.toString(p.peekTop())
	v := arrayGet(array, index)
	p.replaceTop(v)
	return 1, nil
}

func vmArrayLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayIndex := code[i]
	array := p.localArray(int(arrayIndex))
	index := p.toString(p.peekTop())
	v := arrayGet(array, index)
	p.replaceTop(v)
	return 1, nil
}

func vmInGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayIndex := code[i]
	array := p.arrays[arrayIndex]
	index := p.toString(p.peekTop())
	_, ok := array[index]
	p.replaceTop(boolean(ok))
	return 1, nil
}

func vmInLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayIndex := code[i]
	array := p.localArray(int(arrayIndex))
	index := p.toString(p.peekTop())
	_, ok := array[index]
	p.replaceTop(boolean(ok))
	return 1, nil
}

func vmAssignField(p *interp, code []compiler.Opcode, i int) (int, error) {
	right, index := p.popTwo()
	err := p.setField(int(index.num()), p.toString(right))
	if err != nil {
		return 0, err
	}
	return 0, nil
}

func vmAssignGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := code[i]
	p.globals[index] = p.pop()
	return 1, nil
}

func vmAssignLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := code[i]
	p.frame[index] = p.pop()
	return 1, nil
}

func vmAssignSpecial(p *interp, code []compiler.Opcode, i int) (int, error) {
	index := code[i]
	err := p.setSpecial(int(index), p.pop())
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func vmAssignArrayGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayIndex := code[i]
	array := p.arrays[arrayIndex]
	v, index := p.popTwo()
	array[p.toString(index)] = v
	return 1, nil
}

func vmAssignArrayLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayIndex := code[i]
	array := p.localArray(int(arrayIndex))
	v, index := p.popTwo()
	array[p.toString(index)] = v
	return 1, nil
}

func vmDelete(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayScope := code[i]
	arrayIndex := code[i+1]
	array := p.array(ast.VarScope(arrayScope), int(arrayIndex))
	index := p.toString(p.pop())
	delete(array, index)
	return 2, nil
}

func vmDeleteAll(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayScope := code[i]
	arrayIndex := code[i+1]
	array := p.array(ast.VarScope(arrayScope), int(arrayIndex))
	for k := range array {
		delete(array, k)
	}
	return 2, nil
}

func vmIncrField(p *interp, code []compiler.Opcode, i int) (int, error) {
	amount := code[i]
	index := int(p.pop().num())
	v, err := p.getField(index)
	if err != nil {
		return 0, err
	}
	err = p.setField(index, p.toString(num(v.num()+float64(amount))))
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func vmIncrGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	amount := code[i]
	index := code[i+1]
	p.globals[index] = num(p.globals[index].num() + float64(amount))
	return 2, nil
}

func vmIncrLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	amount := code[i]
	index := code[i+1]
	p.frame[index] = num(p.frame[index].num() + float64(amount))
	return 2, nil
}

func vmIncrSpecial(p *interp, code []compiler.Opcode, i int) (int, error) {
	amount := code[i]
	index := int(code[i+1])
	v := p.getSpecial(index)
	err := p.setSpecial(index, num(v.num()+float64(amount)))
	if err != nil {
		return 0, err
	}
	return 2, nil
}

func vmIncrArrayGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	amount := code[i]
	arrayIndex := code[i+1]
	array := p.arrays[arrayIndex]
	index := p.toString(p.pop())
	array[index] = num(array[index].num() + float64(amount))
	return 2, nil
}

func vmIncrArrayLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	amount := code[i]
	arrayIndex := code[i+1]
	array := p.localArray(int(arrayIndex))
	index := p.toString(p.pop())
	array[index] = num(array[index].num() + float64(amount))
	return 2, nil
}

func vmAugAssignField(p *interp, code []compiler.Opcode, i int) (int, error) {
	operation := compiler.AugOp(code[i])
	right, indexVal := p.popTwo()
	index := int(indexVal.num())
	field, err := p.getField(index)
	if err != nil {
		return 0, err
	}
	v, err := p.augAssignOp(operation, field, right)
	if err != nil {
		return 0, err
	}
	err = p.setField(index, p.toString(v))
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func vmAugAssignGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	operation := compiler.AugOp(code[i])
	index := code[i+1]
	v, err := p.augAssignOp(operation, p.globals[index], p.pop())
	if err != nil {
		return 0, err
	}
	p.globals[index] = v
	return 2, nil
}

func vmAugAssignLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	operation := compiler.AugOp(code[i])
	index := code[i+1]
	v, err := p.augAssignOp(operation, p.frame[index], p.pop())
	if err != nil {
		return 0, err
	}
	p.frame[index] = v
	return 2, nil
}

func vmAugAssignSpecial(p *interp, code []compiler.Opcode, i int) (int, error) {
	operation := compiler.AugOp(code[i])
	index := int(code[i+1])
	v, err := p.augAssignOp(operation, p.getSpecial(index), p.pop())
	if err != nil {
		return 0, err
	}
	err = p.setSpecial(index, v)
	if err != nil {
		return 0, err
	}
	return 2, nil
}

func vmAugAssignArrayGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	operation := compiler.AugOp(code[i])
	arrayIndex := code[i+1]
	array := p.arrays[arrayIndex]
	index := p.toString(p.pop())
	v, err := p.augAssignOp(operation, array[index], p.pop())
	if err != nil {
		return 0, err
	}
	array[index] = v
	return 2, nil
}

func vmAugAssignArrayLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	operation := compiler.AugOp(code[i])
	arrayIndex := code[i+1]
	array := p.localArray(int(arrayIndex))
	right, indexVal := p.popTwo()
	index := p.toString(indexVal)
	v, err := p.augAssignOp(operation, array[index], right)
	if err != nil {
		return 0, err
	}
	array[index] = v
	return 2, nil
}

func vmRegex(p *interp, code []compiler.Opcode, i int) (int, error) {
	// Stand-alone /regex/ is equivalent to: $0 ~ /regex/
	index := code[i]
	re := p.regexes[index]
	p.push(boolean(re.MatchString(p.line)))
	return 1, nil
}

func vmMultiIndex(p *interp, code []compiler.Opcode, i int) (int, error) {
	numValues := int(code[i])
	values := p.popSlice(numValues)
	indices := make([]string, 0, 3) // up to 3-dimensional indices won't require heap allocation
	for _, v := range values {
		indices = append(indices, p.toString(v))
	}
	p.push(str(strings.Join(indices, p.subscriptSep)))
	return 1, nil
}

func vmAdd(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	p.replaceTop(num(l.num() + r.num()))
	return 0, nil
}

func vmSubtract(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	p.replaceTop(num(l.num() - r.num()))
	return 0, nil
}

func vmMultiply(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	p.replaceTop(num(l.num() * r.num()))
	return 0, nil
}

func vmDivide(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	rf := r.num()
	if rf == 0.0 {
		return 0, newError("division by zero")
	}
	p.replaceTop(num(l.num() / rf))
	return 0, nil
}

func vmPower(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	p.replaceTop(num(math.Pow(l.num(), r.num())))
	return 0, nil
}

func vmModulo(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	rf := r.num()
	if rf == 0.0 {
		return 0, newError("division by zero in mod")
	}
	p.replaceTop(num(math.Mod(l.num(), rf)))
	return 0, nil
}

func vmEquals(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	if lIsStr || rIsStr {
		p.replaceTop(boolean(p.toString(l) == p.toString(r)))
	} else {
		p.replaceTop(boolean(ln == rn))
	}
	return 0, nil
}

func vmNotEquals(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	if lIsStr || rIsStr {
		p.replaceTop(boolean(p.toString(l) != p.toString(r)))
	} else {
		p.replaceTop(boolean(ln != rn))
	}
	return 0, nil
}

func vmLess(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	if lIsStr || rIsStr {
		p.replaceTop(boolean(p.toString(l) < p.toString(r)))
	} else {
		p.replaceTop(boolean(ln < rn))
	}
	return 0, nil
}

func vmGreater(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	if lIsStr || rIsStr {
		p.replaceTop(boolean(p.toString(l) > p.toString(r)))
	} else {
		p.replaceTop(boolean(ln > rn))
	}
	return 0, nil
}

func vmLessOrEqual(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	if lIsStr || rIsStr {
		p.replaceTop(boolean(p.toString(l) <= p.toString(r)))
	} else {
		p.replaceTop(boolean(ln <= rn))
	}
	return 0, nil
}

func vmGreaterOrEqual(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	if lIsStr || rIsStr {
		p.replaceTop(boolean(p.toString(l) >= p.toString(r)))
	} else {
		p.replaceTop(boolean(ln >= rn))
	}
	return 0, nil
}

func vmConcat(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	p.replaceTop(str(p.toString(l) + p.toString(r)))
	return 0, nil
}

func vmMatch(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	re, err := p.compileRegex(p.toString(r))
	if err != nil {
		return 0, err
	}
	matched := re.MatchString(p.toString(l))
	p.replaceTop(boolean(matched))
	return 0, nil
}

func vmNotMatch(p *interp, code []compiler.Opcode, i int) (int, error) {
	l, r := p.peekPop()
	re, err := p.compileRegex(p.toString(r))
	if err != nil {
		return 0, err
	}
	matched := re.MatchString(p.toString(l))
	p.replaceTop(boolean(!matched))
	return 0, nil
}

func vmNot(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(boolean(!p.peekTop().boolean()))
	return 0, nil
}

func vmUnaryMinus(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(num(-p.peekTop().num()))
	return 0, nil
}

func vmUnaryPlus(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(num(p.peekTop().num()))
	return 0, nil
}

func vmBoolean(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(boolean(p.peekTop().boolean()))
	return 0, nil
}

func vmJump(p *interp, code []compiler.Opcode, i int) (int, error) {
	offset := code[i]
	return 1 + int(offset), nil
}

func vmJumpFalse(p *interp, code []compiler.Opcode, i int) (int, error) {
	offset := code[i]
	v := p.pop()
	if !v.boolean() {
		return 1 + int(offset), nil
	}
	return 1, nil
}

func vmJumpTrue(p *interp, code []compiler.Opcode, i int) (int, error) {
	offset := code[i]
	v := p.pop()
	if v.boolean() {
		return 1 + int(offset), nil
	}
	return 1, nil
}

func vmJumpEquals(p *interp, code []compiler.Opcode, i int) (int, error) {
	offset := code[i]
	l, r := p.popTwo()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	var b bool
	if lIsStr || rIsStr {
		b = p.toString(l) == p.toString(r)
	} else {
		b = ln == rn
	}
	if b {
		return 1 + int(offset), nil
	}
	return 1, nil
}

func vmJumpNotEquals(p *interp, code []compiler.Opcode, i int) (int, error) {
	offset := code[i]
	l, r := p.popTwo()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	var b bool
	if lIsStr || rIsStr {
		b = p.toString(l) != p.toString(r)
	} else {
		b = ln != rn
	}
	if b {
		return 1 + int(offset), nil
	}
	return 1, nil
}

func vmJumpLess(p *interp, code []compiler.Opcode, i int) (int, error) {
	offset := code[i]
	l, r := p.popTwo()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	var b bool
	if lIsStr || rIsStr {
		b = p.toString(l) < p.toString(r)
	} else {
		b = ln < rn
	}
	if b {
		return 1 + int(offset), nil
	}
	return 1, nil
}

func vmJumpGreater(p *interp, code []compiler.Opcode, i int) (int, error) {
	offset := code[i]
	l, r := p.popTwo()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	var b bool
	if lIsStr || rIsStr {
		b = p.toString(l) > p.toString(r)
	} else {
		b = ln > rn
	}
	if b {
		return 1 + int(offset), nil
	}
	return 1, nil
}

func vmJumpLessOrEqual(p *interp, code []compiler.Opcode, i int) (int, error) {
	offset := code[i]
	l, r := p.popTwo()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	var b bool
	if lIsStr || rIsStr {
		b = p.toString(l) <= p.toString(r)
	} else {
		b = ln <= rn
	}
	if b {
		return 1 + int(offset), nil
	}
	return 1, nil
}

func vmJumpGreaterOrEqual(p *interp, code []compiler.Opcode, i int) (int, error) {
	offset := code[i]
	l, r := p.popTwo()
	ln, lIsStr := l.isTrueStr()
	rn, rIsStr := r.isTrueStr()
	var b bool
	if lIsStr || rIsStr {
		b = p.toString(l) >= p.toString(r)
	} else {
		b = ln >= rn
	}
	if b {
		return 1 + int(offset), nil
	}
	return 1, nil
}

func vmNext(p *interp, code []compiler.Opcode, i int) (int, error) {
	return 0, errNext
}

func vmExit(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.exitStatus = int(p.pop().num())
	// Return special errExit value "caught" by top-level executor
	return 0, errExit
}

func vmForIn(p *interp, code []compiler.Opcode, i int) (int, error) {
	varScope := code[i]
	varIndex := code[i+1]
	arrayScope := code[i+2]
	arrayIndex := code[i+3]
	offset := code[i+4]
	array := p.array(ast.VarScope(arrayScope), int(arrayIndex))
	loopCode := code[i+5 : i+5+int(offset)]
	for index := range array {
		switch ast.VarScope(varScope) {
		case ast.ScopeGlobal:
			p.globals[varIndex] = str(index)
		case ast.ScopeLocal:
			p.frame[varIndex] = str(index)
		default: // ScopeSpecial
			err := p.setSpecial(int(varIndex), str(index))
			if err != nil {
				return 0, err
			}
		}
		err := p.execute(loopCode)
		if err == errBreak {
			break
		}
		if err != nil {
			return 0, err
		}
	}
	return 5 + int(offset), nil
}

func vmBreakForIn(p *interp, code []compiler.Opcode, i int) (int, error) {
	return 0, errBreak
}

func vmCallAtan2(p *interp, code []compiler.Opcode, i int) (int, error) {
	y, x := p.peekPop()
	p.replaceTop(num(math.Atan2(y.num(), x.num())))
	return 0, nil
}

func vmCallClose(p *interp, code []compiler.Opcode, i int) (int, error) {
	name := p.toString(p.peekTop())
	var c io.Closer = p.inputStreams[name]
	if c != nil {
		// Close input stream
		delete(p.inputStreams, name)
		err := c.Close()
		if err != nil {
			p.replaceTop(num(-1))
		} else {
			p.replaceTop(num(0))
		}
	} else {
		c = p.outputStreams[name]
		if c != nil {
			// Close output stream
			delete(p.outputStreams, name)
			err := c.Close()
			if err != nil {
				p.replaceTop(num(-1))
			} else {
				p.replaceTop(num(0))
			}
		} else {
			// Nothing to close
			p.replaceTop(num(-1))
		}
	}
	return 0, nil
}

func vmCallCos(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(num(math.Cos(p.peekTop().num())))
	return 0, nil
}

func vmCallExp(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(num(math.Exp(p.peekTop().num())))
	return 0, nil
}

func vmCallFflush(p *interp, code []compiler.Opcode, i int) (int, error) {
	name := p.toString(p.peekTop())
	var ok bool
	if name != "" {
		// Flush a single, named output stream
		ok = p.flushStream(name)
	} else {
		// fflush() or fflush("") flushes all output streams
		ok = p.flushAll()
	}
	if !ok {
		p.replaceTop(num(-1))
	} else {
		p.replaceTop(num(0))
	}
	return 0, nil
}

func vmCallFflushAll(p *interp, code []compiler.Opcode, i int) (int, error) {
	ok := p.flushAll()
	if !ok {
		p.push(num(-1))
	} else {
		p.push(num(0))
	}
	return 0, nil
}

func vmCallGsub(p *interp, code []compiler.Opcode, i int) (int, error) {
	regex, repl, in := p.peekPeekPop()
	out, n, err := p.sub(p.toString(regex), p.toString(repl), p.toString(in), true)
	if err != nil {
		return 0, err
	}
	p.replaceTwo(num(float64(n)), str(out))
	return 0, nil
}

func vmCallIndex(p *interp, code []compiler.Opcode, i int) (int, error) {
	sValue, substr := p.peekPop()
	s := p.toString(sValue)
	index := strings.Index(s, p.toString(substr))
	if p.bytes {
		p.replaceTop(num(float64(index + 1)))
	} else {
		if index < 0 {
			p.replaceTop(num(float64(0)))
		} else {
			index = utf8.RuneCountInString(s[:index])
			p.replaceTop(num(float64(index + 1)))
		}
	}
	return 0, nil
}

func vmCallInt(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(num(float64(int(p.peekTop().num()))))
	return 0, nil
}

func vmCallLength(p *interp, code []compiler.Opcode, i int) (int, error) {
	s := p.line
	var n int
	if p.bytes {
		n = len(s)
	} else {
		n = utf8.RuneCountInString(s)
	}
	p.push(num(float64(n)))
	return 0, nil
}

func vmCallLengthArg(p *interp, code []compiler.Opcode, i int) (int, error) {
	s := p.toString(p.peekTop())
	var n int
	if p.bytes {
		n = len(s)
	} else {
		n = utf8.RuneCountInString(s)
	}
	p.replaceTop(num(float64(n)))
	return 0, nil
}

func vmCallLog(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(num(math.Log(p.peekTop().num())))
	return 0, nil
}

func vmCallMatch(p *interp, code []compiler.Opcode, i int) (int, error) {
	sValue, regex := p.peekPop()
	s := p.toString(sValue)
	re, err := p.compileRegex(p.toString(regex))
	if err != nil {
		return 0, err
	}
	loc := re.FindStringIndex(s)
	if loc == nil {
		p.matchStart = 0
		p.matchLength = -1
		p.replaceTop(num(0))
	} else {
		if p.bytes {
			p.matchStart = loc[0] + 1
			p.matchLength = loc[1] - loc[0]
		} else {
			p.matchStart = utf8.RuneCountInString(s[:loc[0]]) + 1
			p.matchLength = utf8.RuneCountInString(s[loc[0]:loc[1]])
		}
		p.replaceTop(num(float64(p.matchStart)))
	}
	return 0, nil
}

func vmCallRand(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.push(num(p.random.Float64()))
	return 0, nil
}

func vmCallSin(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(num(math.Sin(p.peekTop().num())))
	return 0, nil
}

func vmCallSplit(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayScope := code[i]
	arrayIndex := code[i+1]
	s := p.toString(p.peekTop())
	n, err := p.split(s, ast.VarScope(arrayScope), int(arrayIndex), p.fieldSep)
	if err != nil {
		return 0, err
	}
	p.replaceTop(num(float64(n)))
	return 2, nil
}

func vmCallSplitSep(p *interp, code []compiler.Opcode, i int) (int, error) {
	arrayScope := code[i]
	arrayIndex := code[i+1]
	s, fieldSep := p.peekPop()
	n, err := p.split(p.toString(s), ast.VarScope(arrayScope), int(arrayIndex), p.toString(fieldSep))
	if err != nil {
		return 0, err
	}
	p.replaceTop(num(float64(n)))
	return 2, nil
}

func vmCallSprintf(p *interp, code []compiler.Opcode, i int) (int, error) {
	numArgs := code[i]
	args := p.popSlice(int(numArgs))
	s, err := p.sprintf(p.toString(args[0]), args[1:])
	if err != nil {
		return 0, err
	}
	p.push(str(s))
	return 1, nil
}

func vmCallSqrt(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(num(math.Sqrt(p.peekTop().num())))
	return 0, nil
}

func vmCallSrand(p *interp, code []compiler.Opcode, i int) (int, error) {
	prevSeed := p.randSeed
	p.random.Seed(time.Now().UnixNano())
	p.push(num(prevSeed))
	return 0, nil
}

func vmCallSrandSeed(p *interp, code []compiler.Opcode, i int) (int, error) {
	prevSeed := p.randSeed
	p.randSeed = p.peekTop().num()
	p.random.Seed(int64(math.Float64bits(p.randSeed)))
	p.replaceTop(num(prevSeed))
	return 0, nil
}

func vmCallSub(p *interp, code []compiler.Opcode, i int) (int, error) {
	regex, repl, in := p.peekPeekPop()
	out, n, err := p.sub(p.toString(regex), p.toString(repl), p.toString(in), false)
	if err != nil {
		return 0, err
	}
	p.replaceTwo(num(float64(n)), str(out))
	return 0, nil
}

func vmCallSubstr(p *interp, code []compiler.Opcode, i int) (int, error) {
	sValue, posValue := p.peekPop()
	pos := int(posValue.num())
	s := p.toString(sValue)
	if p.bytes {
		if pos > len(s) {
			pos = len(s) + 1
		}
		if pos < 1 {
			pos = 1
		}
		length := len(s) - pos + 1
		p.replaceTop(str(s[pos-1 : pos-1+length]))
	} else {
		// Count characters till we get to pos.
		chars := 1
		start := 0
		for start = range s {
			chars++
			if chars > pos {
				break
			}
		}
		if pos >= chars {
			start = len(s)
		}

		// Count characters from start till we reach length.
		end := len(s)
		p.replaceTop(str(s[start:end]))
	}
	return 0, nil
}

func vmCallSubstrLength(p *interp, code []compiler.Opcode, i int) (int, error) {
	posValue, lengthValue := p.popTwo()
	length := int(lengthValue.num())
	pos := int(posValue.num())
	s := p.toString(p.peekTop())
	if p.bytes {
		if pos > len(s) {
			pos = len(s) + 1
		}
		if pos < 1 {
			pos = 1
		}
		maxLength := len(s) - pos + 1
		if length < 0 {
			length = 0
		}
		if length > maxLength {
			length = maxLength
		}
		p.replaceTop(str(s[pos-1 : pos-1+length]))
	} else {
		// Count characters till we get to pos.
		chars := 1
		start := 0
		for start = range s {
			chars++
			if chars > pos {
				break
			}
		}
		if pos >= chars {
			start = len(s)
		}

		// Count characters from start till we reach length.
		var end int
		chars = 0
		for end = range s[start:] {
			chars++
			if chars > length {
				break
			}
		}
		if length >= chars {
			end = len(s)
		} else {
			end += start
		}
		p.replaceTop(str(s[start:end]))
	}
	return 0, nil
}

func vmCallSystem(p *interp, code []compiler.Opcode, i int) (int, error) {
	if p.noExec {
		return 0, newError("can't call system() due to NoExec")
	}
	cmdline := p.toString(p.peekTop())
	cmd := p.execShell(cmdline)
	cmd.Stdout = p.output
	cmd.Stderr = p.errorOutput
	_ = p.flushAll() // ensure synchronization
	err := cmd.Start()
	var ret float64
	if err != nil {
		p.printErrorf("%s\n", err)
		ret = -1
	} else {
		err = cmd.Wait()
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				ret = float64(exitErr.ProcessState.ExitCode())
			} else {
				p.printErrorf("unexpected error running command %q: %v\n", cmdline, err)
				ret = -1
			}
		} else {
			ret = 0
		}
	}
	p.replaceTop(num(ret))
	return 0, nil
}

func vmCallTolower(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(str(strings.ToLower(p.toString(p.peekTop()))))
	return 0, nil
}

func vmCallToupper(p *interp, code []compiler.Opcode, i int) (int, error) {
	p.replaceTop(str(strings.ToUpper(p.toString(p.peekTop()))))
	return 0, nil
}

func vmCallUser(p *interp, code []compiler.Opcode, i int) (int, error) {
	funcIndex := code[i]
	numArrayArgs := int(code[i+1])
	i += 2
	numOperands := 2

	f := p.program.Compiled.Functions[funcIndex]
	if p.callDepth >= maxCallDepth {
		return 0, newError("calling %q exceeded maximum call depth of %d", f.Name, maxCallDepth)
	}

	// Set up frame for scalar arguments
	oldFrame := p.frame
	p.frame = p.peekSlice(f.NumScalars)

	// Handle array arguments
	var arrays []int
	for j := 0; j < numArrayArgs; j++ {
		arrayScope := ast.VarScope(code[i])
		arrayIndex := int(code[i+1])
		numOperands += 2
		arrays = append(arrays, p.arrayIndex(arrayScope, arrayIndex))
	}
	oldArraysLen := len(p.arrays)
	for j := numArrayArgs; j < f.NumArrays; j++ {
		arrays = append(arrays, len(p.arrays))
		p.arrays = append(p.arrays, make(map[string]value))
	}
	p.localArrays = append(p.localArrays, arrays)

	// Execute the function!
	p.callDepth++
	err := p.execute(f.Body)
	p.callDepth--

	// Pop the locals off the stack
	p.popSlice(f.NumScalars)
	p.frame = oldFrame
	p.localArrays = p.localArrays[:len(p.localArrays)-1]
	p.arrays = p.arrays[:oldArraysLen]

	if r, ok := err.(returnValue); ok {
		p.push(r.Value)
	} else if err != nil {
		return 0, err
	} else {
		p.push(null())
	}
	return numOperands, nil
}

func vmCallNative(p *interp, code []compiler.Opcode, i int) (int, error) {
	funcIndex := int(code[i])
	numArgs := int(code[i+1])
	args := p.popSlice(numArgs)
	r, err := p.callNative(funcIndex, args)
	if err != nil {
		return 0, err
	}
	p.push(r)
	return 2, nil
}

func vmReturn(p *interp, code []compiler.Opcode, i int) (int, error) {
	v := p.pop()
	return 0, returnValue{v}
}

func vmReturnNull(p *interp, code []compiler.Opcode, i int) (int, error) {
	return 0, returnValue{null()}
}

func vmNulls(p *interp, code []compiler.Opcode, i int) (int, error) {
	numNulls := int(code[i])
	p.pushNulls(numNulls)
	return 1, nil
}

func vmPrint(p *interp, code []compiler.Opcode, i int) (int, error) {
	numArgs := code[i]
	redirect := lexer.Token(code[i+1])

	// Print OFS-separated args followed by ORS (usually newline)
	var line string
	if numArgs > 0 {
		args := p.popSlice(int(numArgs))
		strs := make([]string, len(args))
		for i, a := range args {
			strs[i] = a.str(p.outputFormat)
		}
		line = strings.Join(strs, p.outputFieldSep)
	} else {
		// "print" with no args is equivalent to "print $0"
		line = p.line
	}

	output := p.output
	if redirect != lexer.ILLEGAL {
		var err error
		dest := p.pop()
		output, err = p.getOutputStream(redirect, dest)
		if err != nil {
			return 0, err
		}
	}
	err := p.printLine(output, line)
	if err != nil {
		return 0, err
	}
	return 2, nil
}

func vmPrintf(p *interp, code []compiler.Opcode, i int) (int, error) {
	numArgs := code[i]
	redirect := lexer.Token(code[i+1])

	args := p.popSlice(int(numArgs))
	s, err := p.sprintf(p.toString(args[0]), args[1:])
	if err != nil {
		return 0, err
	}

	output := p.output
	if redirect != lexer.ILLEGAL {
		dest := p.pop()
		output, err = p.getOutputStream(redirect, dest)
		if err != nil {
			return 0, err
		}
	}
	err = writeOutput(output, s)
	if err != nil {
		return 0, err
	}
	return 2, nil
}

func vmGetline(p *interp, code []compiler.Opcode, i int) (int, error) {
	redirect := lexer.Token(code[i])
	ret, line, err := p.getline(redirect)
	if err != nil {
		return 0, err
	}
	if ret == 1 {
		p.setLine(line, false)
	}
	p.push(num(ret))
	return 1, nil
}

func vmGetlineField(p *interp, code []compiler.Opcode, i int) (int, error) {
	redirect := lexer.Token(code[i])
	ret, line, err := p.getline(redirect)
	if err != nil {
		return 0, err
	}
	if ret == 1 {
		err := p.setField(0, line)
		if err != nil {
			return 0, err
		}
	}
	p.push(num(ret))
	return 1, nil
}

func vmGetlineGlobal(p *interp, code []compiler.Opcode, i int) (int, error) {
	redirect := lexer.Token(code[i])
	index := code[i+1]
	ret, line, err := p.getline(redirect)
	if err != nil {
		return 0, err
	}
	if ret == 1 {
		p.globals[index] = numStr(line)
	}
	p.push(num(ret))
	return 2, nil
}

func vmGetlineLocal(p *interp, code []compiler.Opcode, i int) (int, error) {
	redirect := lexer.Token(code[i])
	index := code[i+1]
	ret, line, err := p.getline(redirect)
	if err != nil {
		return 0, err
	}
	if ret == 1 {
		p.frame[index] = numStr(line)
	}
	p.push(num(ret))
	return 2, nil
}

func vmGetlineSpecial(p *interp, code []compiler.Opcode, i int) (int, error) {
	redirect := lexer.Token(code[i])
	index := code[i+1]
	ret, line, err := p.getline(redirect)
	if err != nil {
		return 0, err
	}
	if ret == 1 {
		err := p.setSpecial(int(index), numStr(line))
		if err != nil {
			return 0, err
		}
	}
	p.push(num(ret))
	return 2, nil
}

func vmGetlineArray(p *interp, code []compiler.Opcode, i int) (int, error) {
	redirect := lexer.Token(code[i])
	arrayScope := code[i+1]
	arrayIndex := code[i+2]
	ret, line, err := p.getline(redirect)
	if err != nil {
		return 0, err
	}
	index := p.toString(p.peekTop())
	if ret == 1 {
		array := p.array(ast.VarScope(arrayScope), int(arrayIndex))
		array[index] = numStr(line)
	}
	p.replaceTop(num(ret))
	return 3, nil
}

// Fetch the value at the given index from array. This handles the strange
// POSIX behavior of creating a null entry for non-existent array elements.
// Per the POSIX spec, "Any other reference to a nonexistent array element
// [apart from "in" expressions] shall automatically create it."
func arrayGet(array map[string]value, index string) value {
	v, ok := array[index]
	if !ok {
		array[index] = v
	}
	return v
}

// Stack operations follow. These should be inlined. Instead of just push and
// pop, for efficiency we have custom operations for when we're replacing the
// top of stack without changing the stack pointer. Primarily this avoids the
// check for append in push.
func (p *interp) push(v value) {
	sp := p.sp
	if sp >= len(p.stack) {
		p.stack = append(p.stack, null())
	}
	p.stack[sp] = v
	sp++
	p.sp = sp
}

func (p *interp) pushNulls(num int) {
	sp := p.sp
	for p.sp+num-1 >= len(p.stack) {
		p.stack = append(p.stack, null())
	}
	for i := 0; i < num; i++ {
		p.stack[sp] = null()
		sp++
	}
	p.sp = sp
}

func (p *interp) pop() value {
	p.sp--
	return p.stack[p.sp]
}

func (p *interp) popTwo() (value, value) {
	p.sp -= 2
	return p.stack[p.sp], p.stack[p.sp+1]
}

func (p *interp) peekTop() value {
	return p.stack[p.sp-1]
}

func (p *interp) peekTwo() (value, value) {
	return p.stack[p.sp-2], p.stack[p.sp-1]
}

func (p *interp) peekPop() (value, value) {
	p.sp--
	return p.stack[p.sp-1], p.stack[p.sp]
}

func (p *interp) peekPeekPop() (value, value, value) {
	p.sp--
	return p.stack[p.sp-2], p.stack[p.sp-1], p.stack[p.sp]
}

func (p *interp) replaceTop(v value) {
	p.stack[p.sp-1] = v
}

func (p *interp) replaceTwo(l, r value) {
	p.stack[p.sp-2] = l
	p.stack[p.sp-1] = r
}

func (p *interp) popSlice(n int) []value {
	p.sp -= n
	return p.stack[p.sp : p.sp+n]
}

func (p *interp) peekSlice(n int) []value {
	return p.stack[p.sp-n:]
}

// Helper for getline operations. This performs the (possibly redirected) read
// of a line, and returns the result. If the result is 1 (success in AWK), the
// caller will set the target to the returned string.
func (p *interp) getline(redirect lexer.Token) (float64, string, error) {
	switch redirect {
	case lexer.PIPE: // redirect from command
		name := p.toString(p.pop())
		scanner, err := p.getInputScannerPipe(name)
		if err != nil {
			return 0, "", err
		}
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return -1, "", nil
			}
			return 0, "", nil
		}
		return 1, scanner.Text(), nil

	case lexer.LESS: // redirect from file
		name := p.toString(p.pop())
		scanner, err := p.getInputScannerFile(name)
		if err != nil {
			if _, ok := err.(*os.PathError); ok {
				// File not found is not a hard error, getline just returns -1.
				// See: https://github.com/benhoyt/goawk/issues/41
				return -1, "", nil
			}
			return 0, "", err
		}
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return -1, "", nil
			}
			return 0, "", nil
		}
		return 1, scanner.Text(), nil

	default: // no redirect
		p.flushOutputAndError() // Flush output in case they've written a prompt
		var err error
		line, err := p.nextLine()
		if err == io.EOF {
			return 0, "", nil
		}
		if err != nil {
			return -1, "", nil
		}
		return 1, line, nil
	}
}

// Perform augmented assignment operation.
func (p *interp) augAssignOp(op compiler.AugOp, l, r value) (value, error) {
	switch op {
	case compiler.AugOpAdd:
		return num(l.num() + r.num()), nil
	case compiler.AugOpSub:
		return num(l.num() - r.num()), nil
	case compiler.AugOpMul:
		return num(l.num() * r.num()), nil
	case compiler.AugOpDiv:
		rf := r.num()
		if rf == 0.0 {
			return null(), newError("division by zero")
		}
		return num(l.num() / rf), nil
	case compiler.AugOpPow:
		return num(math.Pow(l.num(), r.num())), nil
	default: // AugOpMod
		rf := r.num()
		if rf == 0.0 {
			return null(), newError("division by zero in mod")
		}
		return num(math.Mod(l.num(), rf)), nil
	}
}
