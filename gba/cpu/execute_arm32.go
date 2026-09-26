package cpu

// TODO handle fetch decode loop for the next instruction after this is executed

func (c *CPU) Branch(opcode uint32) {

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

func (c *CPU) BranchExchange(opcode uint32) {
	//get bits 7-4 to check if BX, BXJ or BLX
	// Don't really need to check this since ARM7TDMI only supports BX
	// Adding in case need to reuse the code for somthing other than GBA

	opType := uint((opcode >> 4) & 0xF)
	// 3-0 denotes the value of Rn
	rn := opcode & 0xF

	switch opType {
	// BX requires 7-4 to be 0001b
	case 0x1:
		c.Registers.Common[PC] = c.Registers.Common[rn]
		// Check if we should switch to THUMB mode
		c.switchModes()
	}

}
