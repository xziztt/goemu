package cpu

func (c *CPU) MoveShiftedRegister(opcode uint16) {
	var (
		rd           = opcode & 0x2
		rs           = (opcode >> 3) & 0x2
		offset       = (opcode >> 3) & 0x5
		op           = (opcode >> 11) & 0x2
		val, shifted uint32
	)

	// op can be LSL, LSR or ASR
	// get value from rs, shift and store into rd
	switch op {
	// LSL -> Shfit left and fill empty pos with 0
	case 0x0:
		val = c.Registers.Common[rs] << offset
		c.Registers.Common[rd] = val
		shifted = (c.Registers.Common[rs] >> (32 - offset)) & 0x1
	// LSR -> shift right, fill empty pos with 0
	case 0x1:
		val = c.Registers.Common[rs] >> offset
		c.Registers.Common[rd] = val
		shifted = (c.Registers.Common[rs] >> (offset - 1)) & 0x1
	// ASR -> arithmetic shift right, calc sign using int32() then shift
	case 0x2:
		val = uint32(int16(c.Registers.Common[rs]) >> offset)
		c.Registers.Common[rd] = val
		shifted = (c.Registers.Common[rs] >> (offset - 1)) & 0x1
	}

	// set flags, Z -> if res is 0, N -> bit 31 of val, C -> last bit shifted out bit
	if val == 0 {
		c.Registers.CPSR |= (1 << Z)
	} else {
		c.Registers.CPSR &^= (1 << Z)
	}

	if (val>>31)&0x1 == 1 {
		c.Registers.CPSR |= (1 << N)
	} else {
		c.Registers.CPSR &^= (1 << N)
	}

	if shifted == 1 {
		c.Registers.CPSR |= (1 << C)
	} else {
		c.Registers.CPSR &^= (1 << N)
	}
}
