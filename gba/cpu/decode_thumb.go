package cpu

func isValidMask(mask, opcode, res uint16) bool {
	return (mask & opcode) == res
}

func isMoveShiftedRegister(opcode uint16) bool {
	return isValidMask(0b1110000000000000, opcode, 0b0000000000000000)
}

func isAddOrSub(opcode uint16) bool {
	return isValidMask(0b1111100000000000, opcode, 0b0001100000000000)
}

func isMCAS(opcode uint16) bool {
	return isValidMask(0b1110000000000000, opcode, 0b0010000000000000)
}

func isALU(opcode uint16) bool {
	return isValidMask(0b1111110000000000, opcode, 0b0100000000000000)
}

func isHiRegisterOps(opcode uint16) bool {
	return isValidMask(0b1111110000000000, opcode, 0b0100010000000000)
}

func isMemoryLoadStore(opcode uint16) bool {
	return isValidMask(0b1111100000000000, opcode, 0b0100100000000000)
}

func isLoadStoreWithRegister(opcode uint16) bool {
	return isValidMask(0b1111000000000000, opcode, 0b0101000000000000)
}

func isLoadStoreSignExtended(opcode uint16) bool {
	return isValidMask(0b1111000000000000, opcode, 0b0101000000000000)
}

func isLoadStoreWithImmOffset(opcode uint16) bool {
	return isValidMask(0b1111000000000000, opcode, 0b0101000000000000)
}

func isLoadStoreHalfWord(opcode uint16) bool {
	return isValidMask(0b1111000000000000, opcode, 0b1000000000000000)
}

func isLoadStoreSPRelative(opcode uint16) bool {
	return isValidMask(0b1111000000000000, opcode, 0b1001000000000000)
}

func isGetRelativeAddress(opcode uint16) bool {
	return isValidMask(0b1111000000000000, opcode, 0b1010000000000000)
}

func isAddOffsetToSp(opcode uint16) bool {
	return isValidMask(0b1111111100000000, opcode, 0b1011000000000000)
}

func isPushPop(opcode uint16) bool {
	return isValidMask(0b1111000000000000, opcode, 0b1011000000000000)
}

func isMultipleLoadStore(opcode uint16) bool {
	return isValidMask(0b1111000000000000, opcode, 0b1100000000000000)
}

func isConditionalBranch(opcode uint16) bool {
	return isValidMask(0b1111000000000000, opcode, 0b1101000000000000)
}

func isUnconditionalBranch(opcode uint16) bool {
	return isValidMask(0b1111100000000000, opcode, 0b1110000000000000)
}

// func isLongBranchWithLink(opcode uint16) bool {
// 	return isValidMask(0b)
// }

func isSoftwareIntrrOrBreakpoint(opcode uint16) bool {
	return isValidMask(0b1111111100000000, opcode, 0b1101111100000000) || isValidMask(0b1111111100000000, opcode, 0b1011111000000000)
}
