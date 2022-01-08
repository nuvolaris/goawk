package bytecode

type Op uint32

const (
	Nop Op = iota

	// Stack operations
	Num   // numIndex
	Str   // strIndex
	Regex // regexIndex
	Dupe
	Drop

	// Fetch a field, variable, or array item
	// TODO: add Field0, Field1, ... FieldN
	Field
	Global      // index
	Local       // index
	Special     // index
	ArrayGlobal // arrayIndex
	ArrayLocal  // arrayIndex
	InGlobal    // arrayIndex
	InLocal     // arrayIndex

	// Assign a field, variable, or array item
	AssignField
	AssignGlobal      // index
	AssignLocal       // index
	AssignSpecial     // index
	AssignArrayGlobal // arrayIndex
	AssignArrayLocal  // arrayIndex
	DeleteGlobal      // arrayIndex
	DeleteLocal       // arrayIndex
	DeleteAllGlobal   // arrayIndex
	DeleteAllLocal    // arrayIndex

	// Post-increment and post-decrement
	PostIncrField
	PostIncrGlobal      // index
	PostIncrLocal       // index
	PostIncrSpecial     // index
	PostIncrArrayGlobal // arrayIndex
	PostIncrArrayLocal  // arrayIndex
	PostDecrField
	PostDecrGlobal      // index
	PostDecrLocal       // index
	PostDecrSpecial     // index
	PostDecrArrayGlobal // arrayIndex
	PostDecrArrayLocal  // arrayIndex

	// Augmented assignment (also used for pre-increment and pre-decrement)
	AugAssignField       // operation
	AugAssignGlobal      // operation, index
	AugAssignLocal       // operation, index
	AugAssignSpecial     // operation, index
	AugAssignArrayGlobal // operation, arrayIndex
	AugAssignArrayLocal  // operation, arrayIndex

	// Binary operators
	Add
	Subtract
	Multiply
	Divide
	Power
	Modulo
	Equals
	NotEquals
	Less
	Greater
	LessOrEqual
	GreaterOrEqual
	Concat
	Match
	NotMatch

	// Unary operators
	Not
	UnaryMinus
	UnaryPlus

	// Control flow
	Jump
	JumpFalse
	JumpTrue
	JumpNumEquals
	JumpNumNotEquals
	JumpNumLess
	JumpNumGreater
	JumpNumLessOrEqual
	JumpNumGreaterOrEqual
	JumpStrEquals
	JumpStrNotEquals
	Return
	Next
	Exit

	// "for (k in a)" combinations
	ForGlobalInGlobal  // varIndex arrayIndex offset
	ForGlobalInLocal   // varIndex arrayIndex offset
	ForLocalInGlobal   // varIndex arrayIndex offset
	ForLocalInLocal    // varIndex arrayIndex offset
	ForSpecialInGlobal // varIndex arrayIndex offset
	ForSpecialInLocal  // varIndex arrayIndex offset

	// Function calls
	// TODO: have separate opcodes for each builtin form, like CallTolower, etc?
	CallBuiltin // func[, numArgs]
	CallUser    // index, numArgs
	CallNative  // index, numArgs

	// Print and getline operations
	Print  // numArgs, redirect
	Printf // numArgs, redirect
	Getline
	GetlineFile
	GetlineCommand
)
