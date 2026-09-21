package arm

func isMatch(opcode uint32, mask uint32, value uint32) bool {
	return (opcode & mask) == value
}

// Branch (B) is supposed to jump to a subroutine. Branch with Link is meant to be used to call to a subroutine, return address is then saved in R14.
//   Bit    Expl.
//   31-28  Condition (must be 1111b for BLX)
//   27-25  Must be "101" for this instruction
//   24     Opcode (0-1) (or Halfword Offset for BLX)
//           0: B{cond} label    ;branch            PC=PC+8+nn*4
//           1: BL{cond} label   ;branch/link       PC=PC+8+nn*4, LR=PC+4
//           H: BLX label ;ARM9  ;branch/link/thumb PC=PC+8+nn*4+H*2, LR=PC+4, T=1
//   23-0   nn - Signed Offset, step 4      (-32M..+32M in steps of 4)

func isBranchInstruction(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1110_0000_0000_0000_0000_0000_0000, 0b0000_1010_0000_0000_0000_0000_0000_0000)
}

// Opcode Format

//   Bit    Expl.
//   31-28  Condition
//   27-8   Must be "0001.0010.1111.1111.1111" for this instruction
//   7-4    Opcode
//           0001b: BX{cond}  Rn    ;PC=Rn, T=Rn.0  (ARMv4T and ARMv5 and up)
//           0011b: BLX{cond} Rn    ;PC=Rn, T=Rn.0, LR=PC+4    (ARMv5 and up)
//   3-0    Rn - Operand Register  (R0-R14)

// Switching to THUMB Mode: Set Bit 0 of the value in Rn to 1, program continues then at Rn-1 in THUMB mode.
// Results in undefined behaviour if using R15 (PC+8 itself) as operand.
// Execution Time: 2S + 1N
// Return: No flags affected.

func isBranchAndExchangeInstruction(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1111_1111_1111_1111_1111_1111_1111, 0b0000_0001_0010_1111_1111_1111_1111_1111)
}
