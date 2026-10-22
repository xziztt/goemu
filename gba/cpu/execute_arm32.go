package cpu

import (
	"math/bits"
)

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
			sign := ((prod >> 31) & 0x1)
			zero := uint32(1)
			if prod == 0 {
				zero = 0
			}
			// clear existing sign bits, set the sign bit using the value of "sign"
			c.Registers.CPSR = CPSR(uint32(c.Registers.CPSR&^(1<<N)) | (sign & (1 << N)))
			c.Registers.CPSR = CPSR(uint32(c.Registers.CPSR&^(1<<Z)) | (zero & (1 << Z)))
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
			sign := 0
			zero := 0
			if prod != 0 {
				zero = 1
			}
			if (prod>>63)&0x1 == 1 {
				sign = 1
			}
			c.Registers.CPSR = CPSR(uint32(c.Registers.CPSR&^(1<<N)) | uint32((sign & (1 << N))))
			c.Registers.CPSR = CPSR(uint32(c.Registers.CPSR&^(1<<Z)) | uint32(zero&(1<<Z)))
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
			zero := 0
			sign := 0
			if (prod>>63)&0x1 == 1 {
				sign = 1
			}
			if prod == 0 {
				zero = 1
			}
			c.Registers.CPSR = CPSR(uint32(c.Registers.CPSR&^(1<<N)) | uint32(sign&(1<<N)))
			c.Registers.CPSR = CPSR(uint32(c.Registers.CPSR&^(1<<Z)) | uint32(zero&(1<<Z)))
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
				carry := (c.Registers.CPSR<<C)&0x1 == 1
				if carry {
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
			// << 3 converts it into bits (x << n = x * 2^n)
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

func (c *CPU) PSRTransfer(opcode uint32) {
	// MSR -> store into PSR, MRS -> load value from PSR into register
	isMsr := (opcode>>21)&0x1 == 1
	if isMsr {
		c.MSR(opcode)
	} else {
		c.MRS(opcode)
	}
}

func (c *CPU) MRS(opcode uint32) {
	spsr := (opcode>>22)&0x1 == 0x1
	rd := (opcode >> 12) & 0xF

	if spsr {
		// Last 5 bits of CPSR denotes the mode
		spsrPos := c.Registers.GetSPSRLoc()
		c.Registers.Common[rd] = c.Registers.SPSR[spsrPos]
	} else {
		c.Registers.Common[rd] = uint32(c.Registers.CPSR)
	}
}

func (c *CPU) MSR(opcode uint32) {
	var value uint32
	r := (opcode >> 22) & 0x1
	immediate := (opcode >> 25) & 0x1
	// r-> 1 = SPSR else CPSR
	// Get mode of the CPU
	// in case of MSR and CPSR combination,
	// 1. we get the current mode of the CPU
	// 2. Read the data in from the source register or immediate addr
	// 3. Write the data into the corresponding index of SPSR (depending on current mode)

	writeMask := uint32(0)
	// bits 16-19 denotes what mask to apply during writes

	if (opcode>>16)&0x1 == 1 {
		writeMask |= 0xFF000000
	}

	if (opcode>>19)&0x1 == 1 {
		writeMask |= 0xFF
	}

	// if immediate the direct value to write will be at 0-8
	if immediate == 1 {
		rotate := ((opcode >> 8) & 0xF) * 2 // Rotate happen in steps of 2 bits
		imm := (opcode & 0xFF)
		// rotate value right
		value = bits.RotateLeft32(imm, -int(rotate))
		// apply mask and write into SPSR loc
		// select all bits other bits than the masked ones, clear out the write masked bits, write only into the cleared bits
	} else {
		rd := opcode & 0xF
		value = c.Registers.Common[rd]
	}

	if r == 1 {
		spsrLoc := c.Registers.GetSPSRLoc()
		c.Registers.SPSR[spsrLoc] = (c.Registers.SPSR[spsrLoc] &^ writeMask) | (value & writeMask)
	} else {
		c.Registers.CPSR = CPSR(uint32(c.Registers.CPSR&^CPSR(writeMask)) | uint32(value&writeMask))
	}
}

func (c *CPU) HalfWord(opcode uint32) {
	const (
		STRH  = 1
		LDRH  = 1
		LDRSB = 2
		LDRSH = 3
	)

	var (
		pre       = opcode >> 24 & 0x1
		up        = opcode >> 23 & 0x1
		imm       = opcode >> 22 & 0x1
		wb        = opcode >> 21 & 0x1
		load      = opcode >> 20 & 0x1
		rn        = opcode >> 16 & 0xF
		rd        = opcode >> 12 & 0xF
		operation = (opcode >> 5) & 0x11
		base      = c.Registers.Common[rn]
		loc       uint32
	)

	var offset uint32
	// if imm flag is set, directly read from bits 0-3 (lower) and 8-11 (upper)
	if imm == 0 {
		offset = (opcode & 0xF) | (((opcode >> 8) & 0xF) << 4)
	} else {
		// Read from offset register
		offset = c.Registers.Common[opcode&0xF]
	}

	// if up, add to base else subtract offset from base
	if up == 1 {
		base += offset
	} else {
		base -= offset
	}

	// pre = 1 -> offset before transfer else before
	if pre == 1 {
		loc = base
	} else {
		loc = c.Registers.Common[rn]
	}

	// Check if store
	if load != 1 {
		switch operation {
		// STRH -> load from Rd store into Rn + offset (if applicable)
		case STRH:
			c.Bus.Write16(loc, uint16(c.Registers.Common[rd]))
		}
	} else {
		switch operation {
		// LDRH Rd, [Rn] -> load the value at address stored in Rn into Rd
		case LDRH:
			// if addr is odd, then ARM7TDMI reads the value and then does a ROR by 8 bits
			if loc&0x1 == 1 {
				val := bits.RotateLeft32(uint32(c.Bus.Read16(loc)), int(-8))
				c.Registers.Common[rd] = val
			}
		// For LDSRB, we take the signed 8 bit value and extend it to 32 bit then store into Rd
		// signed integers are represented in 2s complement form
		case LDRSB:
			val := c.Bus.Read8(loc)
			c.Registers.Common[rd] = uint32(int32(int8(val)))
		// Load 16 bit, sign extend to 32 bit and load into rd
		case LDRSH:
			val := c.Bus.Read16(loc)
			c.Registers.Common[rd] = uint32(int32(int16(val)))
		}
	}

	// write back calculated addr into rn
	if wb == 1 || pre != 1 {
		c.Registers.Common[rn] = loc
	}
}

func (c *CPU) SignedDataTransfer(opcode uint32) {
	// TODO: Implement
}

func (c *CPU) Block(opcode uint32) {
	// TODO: Implement
}

func (c *CPU) SingleDataSwap(opcode uint32) {
	// TODO: Implement
}

func (c *CPU) DataProcessing(opcode uint32) {
	// TODO: Implement
}
