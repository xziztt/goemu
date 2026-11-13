package cpu

import (
	"math/bits"
)

func (c *CPU) MoveShiftedRegister(opcode uint16) {
	var (
		rd           = opcode & 0x7
		rs           = (opcode >> 3) & 0x7
		offset       = (opcode >> 6) & 0x1F
		op           = (opcode >> 11) & 0x3
		val, shifted uint32
	)

	// op can be LSL, LSR or ASR
	// get value from rs, shift and store into rd
	switch op {
	// LSL -> Shfit left and fill empty pos with 0
	case 0x0:
		val = (c.Registers.Common[rs] << offset)
		shifted = (c.Registers.Common[rs] >> (32 - offset)) & 0x1
		c.Registers.Common[rd] = val
	// LSR -> shift right, fill empty pos with 0
	case 0x1:
		// LSR#0/ASR#0 are interpreted as LSR#32/ASR#32
		if offset == 0 {
			offset = 32
		}
		val = (c.Registers.Common[rs] >> offset)
		shifted = (c.Registers.Common[rs] >> (offset - 1)) & 0x1
		c.Registers.Common[rd] = val
	// ASR -> arithmetic shift right, calc sign using int32() then shift
	case 0x2:
		if offset == 0 {
			offset = 32
		}
		val = uint32(int32(c.Registers.Common[rs]) >> offset)
		shifted = (c.Registers.Common[rs] >> (offset - 1)) & 0x1
		c.Registers.Common[rd] = val
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

	if offset != 0 {
		if shifted == 1 {
			c.Registers.CPSR |= (1 << C)
		} else {
			c.Registers.CPSR &^= (1 << C)
		}
	}
}

func (c *CPU) AddSub(opcode uint16) {
	var (
		rd                        = opcode & 0x7
		rs                        = (opcode >> 3) & 0x7
		rn                        = (opcode >> 6) & 0x7
		nn                        = (opcode >> 6) & 0x7
		op                        = (opcode >> 9) & 0x3
		sum                       uint64
		val                       uint32
		carry                     uint32
		v                         uint32
		op1, op2                  uint32
		signop1, signop2, signRes uint32
	)

	signop1 = (c.Registers.Common[rs] >> 31) & 0x1
	op1 = c.Registers.Common[rs]
	switch op {
	// Add register
	case 0x0:
		sum = uint64(c.Registers.Common[rs]) + uint64(c.Registers.Common[rn])
		val = uint32(sum)
		signop2 = (c.Registers.Common[rn] >> 31) & 0x1
		c.Registers.Common[rd] = val
		if sum>>32 == 0x1 {
			carry = 1
		} else {
			carry = 0
		}
	case 0x1:
		op2 = c.Registers.Common[rn]
		val = uint32(int32(op1 - op2))
		c.Registers.Common[rd] = val
		signop2 = (op2 >> 31) & 0x1
		if op1 >= op2 {
			carry = 1
		} else {
			carry = 0
		}
		// for immediate, nn is 3 bits, hence when interpreted as signed 32 bit, it will always be pos
		// so no need to consider for V calc since the sign is always pos
	case 0x2:
		op264 := uint64(nn)
		sum = uint64(op1) + op264
		val = uint32(sum)
		c.Registers.Common[rd] = val
		signop2 = 0
		if sum>>32 == 0x1 {
			carry = 1
		} else {
			carry = 0
		}
	case 0x3:
		op2 = uint32(nn)
		val = op1 - op2
		c.Registers.Common[rd] = val
		signop2 = 0
		if op1 >= op2 {
			carry = 1
		} else {
			carry = 0
		}
	}

	// N,Z,C and V are affected
	if val == 0 {
		c.Registers.SetCPSRBit(Z, 1)
	} else {
		c.Registers.SetCPSRBit(Z, 0)
	}

	// for add C = carry past 31st bit, for sub c = 0 if borrow else c = 1
	if carry == 1 {
		c.Registers.SetCPSRBit(C, 1)
	} else {
		c.Registers.SetCPSRBit(C, 0)
	}

	if (val>>31)&0x1 == 1 {
		c.Registers.SetCPSRBit(N, 1)
	} else {
		c.Registers.SetCPSRBit(N, 0)
	}

	signRes = (val >> 31) & 0x1

	switch op {
	// in case of add, v = 1 in 2 cases -> op1 and op2 both pos and res neg, op1 and op2 both neg and res pos
	case 0x0, 0x2:
		if (signop1 == signop2) && (signRes != signop1) {
			v = 1
		} else {
			v = 0
		}
	// for sub, v = 1 cases -> op1 is neg, op2 is pos, res is pos. op1 is pos, op2 is neg, res is neg
	case 0x1, 0x3:
		if (signop1 != signop2) && (signRes != signop1) {
			v = 1
		} else {
			v = 0
		}
	}
	if v == 1 {
		c.Registers.SetCPSRBit(V, 1)
	} else {
		c.Registers.SetCPSRBit(V, 0)
	}
}

func (c *CPU) MovCmpAddSubImm(opcode uint16) {
	var (
		op                      = (opcode >> 9) & 0x2
		rd                      = (opcode & 0x7)
		nn                      = opcode & 0x7F
		val, carry, overflow, z uint32
		signOp1, signRes        uint32
		op1, op2                uint32
	)
	op1 = c.Registers.Common[rd]
	op2 = uint32(nn)

	signOp1 = (c.Registers.Common[rd] >> 31) & 0x1
	switch op {
	case 0x0:
		val = op2
		c.Registers.Common[rd] = uint32(nn)
		if nn == 0 {
			z = 1
		} else {
			z = 0
		}
	case 0x1:
		val = op1 - op2
		if rd == nn {
			z = 1
		} else {
			z = 0
		}

		if op1 >= op2 {
			carry = 1
		} else {
			carry = 0
		}

		signRes = (val >> 31) & 0x1
		if signOp1 == 1 && signRes == 0 {
			overflow = 1
		}

	case 0x2:
		val64 := uint64(op1) + uint64(op2)
		val = uint32(val64)
		c.Registers.Common[rd] = val
		if val == 0 {
			z = 1
		} else {
			z = 0
		}
		signRes = (val >> 31) & 0x1
		if (signOp1 == 0 && signRes == 1) || (signOp1 == 1 && signRes == 0) {
			overflow = 1
		} else {
			overflow = 0
		}
		// For add operation, if res does not fit in the 32 bit int var, then c = 1
		if uint32(val64>>32)&0x1 == 1 {
			carry = 1
		} else {
			carry = 0
		}
	case 0x3:
		val = c.Registers.Common[rd] - uint32(nn)
		c.Registers.Common[rd] = val
		if val == 0 {
			z = 1
		} else {
			z = 0
		}
		if (signOp1 == 0 && signRes == 1) || (signOp1 == 1 && signRes == 0) {
			overflow = 1
		} else {
			overflow = 0
		}

		// for sub, if res >=0, that means no borrow, then c = 1 else 0
		if op1 >= op2 {
			carry = 1
		} else {
			carry = 0
		}
	}

	// N is the 31st bit of the value
	if (val>>31)&0x1 == 1 {
		c.Registers.SetCPSRBit(N, 1)
	} else {
		c.Registers.SetCPSRBit(N, 0)
	}

	c.Registers.SetCPSRBit(C, carry)
	c.Registers.SetCPSRBit(V, overflow)
	c.Registers.SetCPSRBit(Z, z)
}

func (c *CPU) ALUOperations(opcode uint16) {
	var (
		rd                        = opcode & 0x7
		rs                        = (opcode >> 3) & 0x7
		op                        = (opcode >> 6) & 0xF
		carry, v, op1, op2, val   uint32
		op1Sign, op2Sign, resSign uint32
	)

	op1 = c.Registers.Common[rd]
	op2 = c.Registers.Common[rs]
	op1Sign = (op1 >> 31) & 0x1
	op2Sign = (op2 >> 31) & 0x1

	switch op {
	// ADD
	case 0x0:
		val64 := uint64(op1) + uint64(op2)
		val = uint32(val64)
		if (val64>>32)&0x1 == 1 {
			carry = 1
		} else {
			carry = 0
		}
		resSign = (val >> 31) & 0x1
		if op1Sign == op2Sign && op1Sign != resSign {
			v = 1
		} else {
			v = 0
		}
		c.Registers.Common[rd] = val

	// EOR
	case 0x1:
		val = op1 ^ op2
		c.Registers.Common[rd] = val
	// LSL
	case 0x2:
		rv := (op2 & 0xFF)
		val = op1 << rv
		c.Registers.Common[rd] = val
		// C is the last shifted out bit
		if rv != 0 {
			carry = (op1 >> (32 - rv)) & 0x1
		}
	// LSR
	case 0x3:
		rv := (op2 & 0xFF)
		val = op1 >> rv
		c.Registers.Common[rd] = val
		if rv != 0 {
			carry = (op1 >> (rv - 1)) & 0x1
		}

	// ASR
	case 0x4:
		rv := uint32(int32(op2 & 0xFF))
		val = op1 >> rv
		c.Registers.Common[rd] = val
		if rv != 0 {
			carry = (op1 >> (rv - 1)) & 0x1
		}
	// ADC
	case 0x5:
		val64 := uint64(op1) + uint64(op2) + uint64(c.Registers.GetCPSRBit(C))
		val = uint32(val64)
		c.Registers.Common[rd] = val

		if (val64>>32)&0x1 == 1 {
			carry = 1
		} else {
			carry = 0
		}
		resSign = (val >> 31) & 0x1
		if op1Sign == op2Sign && op1Sign != resSign {
			v = 1
		} else {
			v = 0
		}
	// SBC
	case 0x6:
		val = op1 - op2 - c.Registers.GetCPSRBit(C)
		c.Registers.Common[rd] = val
		if op1 >= op2 {
			carry = 1
		} else {
			carry = 0
		}

		resSign = (val >> 31) * 0x1
		if op1Sign != op2Sign && op1Sign != resSign {
			v = 1
		} else {
			v = 0
		}

	// ROR
	case 0x7:
		rv := op2 & 0xFF
		val = bits.RotateLeft32(op1, -int(rv))
		// if rv != 0, set to 31st bit of rotated value
		if rv != 0 {
			carry = (val >> 31) & 0x1
		}
		c.Registers.Common[rd] = val
	// TST
	case 0x8:
		val = op1 & op2
	// NEG
	case 0x9:
		val = 0 - op2
		c.Registers.Common[rd] = val
		if op2 == 0 {
			carry = 1
		} else {
			carry = 0
		}
		resSign = (val >> 31) & 0x1
		if op1Sign != op2Sign && op1Sign != resSign {
			v = 1
		} else {
			v = 0
		}

	// CMP
	case 0xA:
		val = op1 - op2
		if op1 >= op2 {
			carry = 1
		} else {
			carry = 0
		}
		resSign = (val >> 31) & 0x1
		if op1Sign != op2Sign && op1Sign != resSign {
			v = 1
		} else {
			v = 0
		}

	// CMN
	case 0xB:
		val64 := uint64(op1) + uint64(op2)
		val = uint32(val64)
		if (val64>>32)&0x1 == 1 {
			carry = 1
		} else {
			carry = 0
		}
		resSign = (val >> 31) & 0x1
		if op1Sign == op2Sign && op1Sign != resSign {
			v = 1
		} else {
			v = 0
		}
	// ORR
	case 0xC:
		val = op1 | op2
		c.Registers.Common[rd] = val
	// MUL
	case 0xD:
		val = op1 * op2
		c.Registers.Common[rd] = val
	// BIC
	case 0xE:
		val = op1 &^ op2
		c.Registers.Common[rd] = val
	// MVN
	case 0xF:
		val = ^op2
		c.Registers.Common[rd] = val
	}

	if val == 0 {
		c.Registers.SetCPSRBit(Z, 1)
	} else {
		c.Registers.SetCPSRBit(Z, 0)
	}

	// 31st bit is N
	if (val>>31)&0x1 == 1 {
		c.Registers.SetCPSRBit(N, 1)
	} else {
		c.Registers.SetCPSRBit(N, 0)
	}
	c.Registers.SetCPSRBit(C, carry)
	c.Registers.SetCPSRBit(V, v)
}
