// Code generated by "stringer -type=Opcode"; DO NOT EDIT.

package compiler

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Nop-0]
	_ = x[Num-1]
	_ = x[Str-2]
	_ = x[Dupe-3]
	_ = x[Drop-4]
	_ = x[Swap-5]
	_ = x[Field-6]
	_ = x[FieldNum-7]
	_ = x[Global-8]
	_ = x[Local-9]
	_ = x[Special-10]
	_ = x[ArrayGlobal-11]
	_ = x[ArrayLocal-12]
	_ = x[InGlobal-13]
	_ = x[InLocal-14]
	_ = x[AssignField-15]
	_ = x[AssignGlobal-16]
	_ = x[AssignLocal-17]
	_ = x[AssignSpecial-18]
	_ = x[AssignArrayGlobal-19]
	_ = x[AssignArrayLocal-20]
	_ = x[DeleteGlobal-21]
	_ = x[DeleteLocal-22]
	_ = x[DeleteAllGlobal-23]
	_ = x[DeleteAllLocal-24]
	_ = x[IncrField-25]
	_ = x[IncrGlobal-26]
	_ = x[IncrLocal-27]
	_ = x[IncrSpecial-28]
	_ = x[IncrArrayGlobal-29]
	_ = x[IncrArrayLocal-30]
	_ = x[AugAssignField-31]
	_ = x[AugAssignGlobal-32]
	_ = x[AugAssignLocal-33]
	_ = x[AugAssignSpecial-34]
	_ = x[AugAssignArrayGlobal-35]
	_ = x[AugAssignArrayLocal-36]
	_ = x[Regex-37]
	_ = x[MultiIndex-38]
	_ = x[Add-39]
	_ = x[Subtract-40]
	_ = x[Multiply-41]
	_ = x[Divide-42]
	_ = x[Power-43]
	_ = x[Modulo-44]
	_ = x[Equals-45]
	_ = x[NotEquals-46]
	_ = x[Less-47]
	_ = x[Greater-48]
	_ = x[LessOrEqual-49]
	_ = x[GreaterOrEqual-50]
	_ = x[Concat-51]
	_ = x[Match-52]
	_ = x[NotMatch-53]
	_ = x[Not-54]
	_ = x[UnaryMinus-55]
	_ = x[UnaryPlus-56]
	_ = x[Boolean-57]
	_ = x[Jump-58]
	_ = x[JumpFalse-59]
	_ = x[JumpTrue-60]
	_ = x[JumpEquals-61]
	_ = x[JumpNotEquals-62]
	_ = x[JumpLess-63]
	_ = x[JumpGreater-64]
	_ = x[JumpLessOrEqual-65]
	_ = x[JumpGreaterOrEqual-66]
	_ = x[Next-67]
	_ = x[Exit-68]
	_ = x[ForInGlobal-69]
	_ = x[ForInLocal-70]
	_ = x[ForInSpecial-71]
	_ = x[BreakForIn-72]
	_ = x[CallAtan2-73]
	_ = x[CallClose-74]
	_ = x[CallCos-75]
	_ = x[CallExp-76]
	_ = x[CallFflush-77]
	_ = x[CallFflushAll-78]
	_ = x[CallGsub-79]
	_ = x[CallIndex-80]
	_ = x[CallInt-81]
	_ = x[CallLength-82]
	_ = x[CallLengthArg-83]
	_ = x[CallLog-84]
	_ = x[CallMatch-85]
	_ = x[CallRand-86]
	_ = x[CallSin-87]
	_ = x[CallSplitGlobal-88]
	_ = x[CallSplitLocal-89]
	_ = x[CallSplitSepGlobal-90]
	_ = x[CallSplitSepLocal-91]
	_ = x[CallSprintf-92]
	_ = x[CallSqrt-93]
	_ = x[CallSrand-94]
	_ = x[CallSrandSeed-95]
	_ = x[CallSub-96]
	_ = x[CallSubstr-97]
	_ = x[CallSubstrLength-98]
	_ = x[CallSystem-99]
	_ = x[CallTolower-100]
	_ = x[CallToupper-101]
	_ = x[CallUser-102]
	_ = x[CallNative-103]
	_ = x[Return-104]
	_ = x[ReturnNull-105]
	_ = x[Nulls-106]
	_ = x[Print-107]
	_ = x[Printf-108]
	_ = x[Getline-109]
	_ = x[GetlineField-110]
	_ = x[GetlineGlobal-111]
	_ = x[GetlineLocal-112]
	_ = x[GetlineSpecial-113]
	_ = x[GetlineArrayGlobal-114]
	_ = x[GetlineArrayLocal-115]
}

const _Opcode_name = "NopNumStrDupeDropSwapFieldFieldNumGlobalLocalSpecialArrayGlobalArrayLocalInGlobalInLocalAssignFieldAssignGlobalAssignLocalAssignSpecialAssignArrayGlobalAssignArrayLocalDeleteGlobalDeleteLocalDeleteAllGlobalDeleteAllLocalIncrFieldIncrGlobalIncrLocalIncrSpecialIncrArrayGlobalIncrArrayLocalAugAssignFieldAugAssignGlobalAugAssignLocalAugAssignSpecialAugAssignArrayGlobalAugAssignArrayLocalRegexMultiIndexAddSubtractMultiplyDividePowerModuloEqualsNotEqualsLessGreaterLessOrEqualGreaterOrEqualConcatMatchNotMatchNotUnaryMinusUnaryPlusBooleanJumpJumpFalseJumpTrueJumpEqualsJumpNotEqualsJumpLessJumpGreaterJumpLessOrEqualJumpGreaterOrEqualNextExitForInGlobalForInLocalForInSpecialBreakForInCallAtan2CallCloseCallCosCallExpCallFflushCallFflushAllCallGsubCallIndexCallIntCallLengthCallLengthArgCallLogCallMatchCallRandCallSinCallSplitGlobalCallSplitLocalCallSplitSepGlobalCallSplitSepLocalCallSprintfCallSqrtCallSrandCallSrandSeedCallSubCallSubstrCallSubstrLengthCallSystemCallTolowerCallToupperCallUserCallNativeReturnReturnNullNullsPrintPrintfGetlineGetlineFieldGetlineGlobalGetlineLocalGetlineSpecialGetlineArrayGlobalGetlineArrayLocal"

var _Opcode_index = [...]uint16{0, 3, 6, 9, 13, 17, 21, 26, 34, 40, 45, 52, 63, 73, 81, 88, 99, 111, 122, 135, 152, 168, 180, 191, 206, 220, 229, 239, 248, 259, 274, 288, 302, 317, 331, 347, 367, 386, 391, 401, 404, 412, 420, 426, 431, 437, 443, 452, 456, 463, 474, 488, 494, 499, 507, 510, 520, 529, 536, 540, 549, 557, 567, 580, 588, 599, 614, 632, 636, 640, 651, 661, 673, 683, 692, 701, 708, 715, 725, 738, 746, 755, 762, 772, 785, 792, 801, 809, 816, 831, 845, 863, 880, 891, 899, 908, 921, 928, 938, 954, 964, 975, 986, 994, 1004, 1010, 1020, 1025, 1030, 1036, 1043, 1055, 1068, 1080, 1094, 1112, 1129}

func (i Opcode) String() string {
	if i < 0 || i >= Opcode(len(_Opcode_index)-1) {
		return "Opcode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Opcode_name[_Opcode_index[i]:_Opcode_index[i+1]]
}
