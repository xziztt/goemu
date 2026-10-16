package cpu

import "math/bits"

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
	// get bits 7-4 to check if BX, BXJ or BLX
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

func (c *CPU) SingleDataTransfer(opcode uint32) {
	const (
		LSL = iota
		LSR
		ASR
		ROR
	)

	var (
		sr   = opcode>>25&0x1 == 1
		pre  = opcode>>24&0x1 == 1
		up   = opcode>>23&0x1 == 1
		byte = opcode>>22&0x1 == 1
		//wb   = opcode>>21&0x1 == 1 && !pre
		load = opcode>>20&0x1 == 1
		rn   = opcode >> 16 & 0xF
		rd   = opcode >> 12 & 0xF
		addr uint32
	)

	var offset uint32

	// if !sr, bits 11-0 contains the offset
	if !sr {
		offset = opcode & 0xFFF
	} else {
		// offset register number
		or := opcode & 0xF
		shiftType := (opcode >> 5) & 3
		shiftAmount := (opcode >> 7) & 31

		//   Zero Shift Amount (Shift Register by Immediate, with Immediate=0)
		//   LSL#0: No shift performed, ie. directly Op2=Rm, the C flag is NOT affected.
		//   LSR#0: Interpreted as LSR#32, ie. Op2 becomes zero, C becomes Bit 31 of Rm.
		//   ASR#0: Interpreted as ASR#32, ie. Op2 and C are filled by Bit 31 of Rm.
		//   ROR#0: Interpreted as RRX#1 (RCR), like ROR#1, but Op2 Bit 31 set to old C.

		switch shiftType {
		case LSL:
			offset = c.Registers.Common[or] << shiftAmount
		// Logical shift right
		case LSR:
			if shiftAmount == 0 {
				shiftAmount = 32
			}
			offset = c.Registers.Common[or] >> shiftAmount
		// Arithmetic Shift Right
		case ASR:
			if shiftAmount == 0 {
				shiftAmount = 32
			}
			offset = uint32(c.Registers.Common[or] >> shiftAmount)
		// Rotate Right
		case ROR:
			// if shift is for 32, it is RRX which means rotate right with carry flag participating
			if shiftAmount == 0 {
				// shift everything by 1
				offset = uint32(c.Registers.Common[or] >> 1)
				// if carry bit is set
				if c.Registers.CPSR.C {
					// set 31st bit of offset as 1 (carry should rotate right)
					offset = offset | 0x80000000
				}

			} else {
				// else do a normal shift to the right by shiftAmount
				offset = bits.RotateLeft32(c.Registers.Common[or], -int(shiftAmount))
			}
		}

	}

	// base address
	loc := c.Registers.Common[rn]

	if up {
		loc += offset
	} else {
		loc -= offset
	}

	// 1 -> PRE -> add offset before tansfer, 0 -> POST -> add offset after transfer
	// PRE means -> Access memory[current + offset], POST means -> Access memory[current], update address to current + offset

	if pre {
		addr = loc
	} else {
		addr = c.Registers.Common[rn]
	}

	// check the load bit, if load bit is 1, load from memory, else store
	if load {
		// Check byte bit, if byte bit is 1, then transfer 8 bit else 32 bit
		if byte {
			// Transfer 1 byte
			c.Registers.Common[rd] = uint32(c.Bus.Read8(addr))
		} else {
			// if transferring 32 bits, first read 32 bits from the addr
			// figure out how much to shift by so that the value at addr is LSB of the word aligned value
			val := uint32(c.Bus.Read32(addr))
			// val & 3 gives you the offset from word aligned addr (word aligned = divisible by 4)
			// << 3 converts it into bits
			rotateBy := (addr & 3) << 3
			c.Registers.Common[rd] = bits.RotateLeft32(val, -int(rotateBy))
		}
	} else {
		// store
		val := c.Registers.Common[rd]
		if byte {
			c.Bus.Write8(addr, uint8(val))
		} else {
			c.Bus.Write32(addr, val)
		}
	}
}
