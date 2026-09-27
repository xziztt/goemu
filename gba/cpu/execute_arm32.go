package cpu

// TODO handle fetch decode loop for the next instruction after this is executed

func (c *CPU) getCyclesForMul(rs uint32) uint32 {
	mask := uint32(0xFFFFFF00)
	cycles := uint32(0)
	for {
		checkBits := rs & mask
		if checkBits == 0 || checkBits == mask {
			break
		}
		mask = mask << 8
		cycles += 1
	}

	return cycles

}

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

func (c *CPU) Multiply(opcode uint32) {

	const (
		MUL   = 0b0000
		MLA   = 0b0001
		UMULL = 0b0100
		UMLAL = 0b0101
		SMULL = 0b0110
		SMLAL = 0b0111
	)

	set := ((opcode >> 20) & 0x1) != 0
	rd := (opcode >> 16) & 0xF
	rn := (opcode >> 12) & 0xF
	rs := (opcode >> 8) & 0xF
	rm := opcode & 0xF

	// get bits starting from 21-24

	instr := (opcode >> 21) & 0xF
	switch instr {
	case MUL, MLA:
		prod := c.Registers.Common[rm] * c.Registers.Common[rs]

		// if MLA add value from rn
		if instr == MLA {
			prod += c.Registers.Common[rn]
		}

		c.Registers.Common[rd] = prod

		// ToDo calculate cycles and then update timings accordingly
		cycles := c.getCyclesForMul(c.Registers.Common[rs])
		c.skip(cycles)

		// if set is 1, then set Zero Flag and the Sign Flag,
		if set {
			c.Registers.CPSR.N = (opcode>>31)&0x1 == 1
			c.Registers.CPSR.Z = prod == 0
		}

	case UMULL, UMLAL:
		prod := uint64(c.Registers.Common[rm]) * uint64(c.Registers.Common[rs])
		if instr == UMLAL {
			// if UMLAL, get first 32 bits from rdHI, get second 32 from rdLo and concat them
			prod += uint64(c.Registers.Common[rd])<<32 | uint64(c.Registers.Common[rn])
		}

		c.Registers.Common[rd] = uint32(prod >> 32)
		c.Registers.Common[rn] = uint32(prod)

		cycles := c.getCyclesForMul(c.Registers.Common[rs])
		c.skip(cycles)

		if set {
			c.Registers.CPSR.N = (prod>>63)&0x1 == 1
			c.Registers.CPSR.Z = prod == 0

		}

	case SMULL, SMLAL:
		prod := int64(c.Registers.Common[rm]) * int64(c.Registers.Common[rs])
		if instr == SMLAL {
			prod += int64(c.Registers.Common[rd])<<32 | int64(c.Registers.Common[rn])
		}

		c.Registers.Common[rd] = uint32(prod >> 32)
		c.Registers.Common[rn] = uint32(prod)

		cycles := c.getCyclesForMul(c.Registers.Common[rs])
		c.skip(cycles)

		if set {
			c.Registers.CPSR.N = (prod>>63)&0x1 == 1
			c.Registers.CPSR.Z = prod == 0
		}

	}
}
