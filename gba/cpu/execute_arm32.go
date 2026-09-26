package cpu

func (c *CPU) ExecBranch(opcode uint32) {

	// bit 24 - if 0 -> B, if 1 -> BL

	// if BL, store the return addr (PC - 4) inside LR
	if (opcode>>24)&1 == 1 {
		c.Registers.Common[LR] = c.Registers.Common[PC] - 4
	}

	// 3-0   nn - Signed Offset , step 4      (-32M..+32M in steps of 4)
	// This tells how many steps of 4 bytes should be taken to reach
	// store the last 24 bits which represent nn, which means there are nn * 4 bytes from the current PC to jump to
	// (PC + 8 + 4*nn basically)
	// ARM follows a decode -> fetch -> exec pipeline, so if current isntr is in execute phase, that means pc is at +8 already

	// opcode << 8 gives you last 24 bits, now right shift these by 8 to give you signed representation, left shift by 2 to multiple nn by 4 (>>8<<4 = >> 6)
	newLoc := uint32((opcode << 8) >> 6)
	c.Registers.Common[PC] = newLoc
}
