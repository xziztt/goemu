package cpu

import (
	"fmt"
	"testing"
)

func TestMoveShiftedRegister(t *testing.T) {
	tests := []struct {
		name      string
		operation uint16 // 0 = LSL, 1 = LSR, 2 = ASR
		amount    uint16 // Encoded 5-bit immediate
		input     uint32
		initialC  uint32
		want      uint32
		wantN     uint32
		wantZ     uint32
		wantC     uint32
	}{
		{
			name:      "LSL zero preserves set carry",
			operation: 0,
			amount:    0,
			input:     0x12345678,
			initialC:  1,
			want:      0x12345678,
			wantC:     1,
		},
		{
			name:      "LSL zero preserves clear carry",
			operation: 0,
			amount:    0,
			input:     0x80000000,
			want:      0x80000000,
			wantN:     1,
		},
		{
			name:      "LSL zero input sets Z",
			operation: 0,
			amount:    0,
			input:     0,
			want:      0,
			wantZ:     1,
		},
		{
			name:      "LSL one sets carry from bit 31",
			operation: 0,
			amount:    1,
			input:     0x80000001,
			want:      0x00000002,
			wantC:     1,
		},
		{
			name:      "LSL one sets N and clears carry",
			operation: 0,
			amount:    1,
			input:     0x40000000,
			initialC:  1,
			want:      0x80000000,
			wantN:     1,
		},
		{
			name:      "LSL shifts all set bits out",
			operation: 0,
			amount:    1,
			input:     0x80000000,
			want:      0,
			wantZ:     1,
			wantC:     1,
		},
		{
			name:      "LSL 31 sets carry from original bit 1",
			operation: 0,
			amount:    31,
			input:     0x00000003,
			want:      0x80000000,
			wantN:     1,
			wantC:     1,
		},
		{
			name:      "LSR one sets carry from bit 0",
			operation: 1,
			amount:    1,
			input:     0x80000001,
			want:      0x40000000,
			wantC:     1,
		},
		{
			name:      "LSR one clears carry",
			operation: 1,
			amount:    1,
			input:     0x80000002,
			initialC:  1,
			want:      0x40000001,
		},
		{
			name:      "LSR 31 sets carry from original bit 30",
			operation: 1,
			amount:    31,
			input:     0xC0000000,
			want:      1,
			wantC:     1,
		},
		{
			name:      "LSR encoded zero means 32 with carry set",
			operation: 1,
			amount:    0,
			input:     0x80000000,
			want:      0,
			wantZ:     1,
			wantC:     1,
		},
		{
			name:      "LSR encoded zero means 32 with carry clear",
			operation: 1,
			amount:    0,
			input:     0x7FFFFFFF,
			initialC:  1,
			want:      0,
			wantZ:     1,
		},
		{
			name:      "ASR one preserves negative sign",
			operation: 2,
			amount:    1,
			input:     0x80000001,
			want:      0xC0000000,
			wantN:     1,
			wantC:     1,
		},
		{
			name:      "ASR one clears carry",
			operation: 2,
			amount:    1,
			input:     0x80000002,
			initialC:  1,
			want:      0xC0000001,
			wantN:     1,
		},
		{
			name:      "ASR positive one becomes zero",
			operation: 2,
			amount:    1,
			input:     1,
			want:      0,
			wantZ:     1,
			wantC:     1,
		},
		{
			name:      "ASR 31 uses original bit 30 for carry",
			operation: 2,
			amount:    31,
			input:     0x80000000,
			initialC:  1,
			want:      0xFFFFFFFF,
			wantN:     1,
		},
		{
			name:      "ASR encoded zero means 32 for negative input",
			operation: 2,
			amount:    0,
			input:     0x80000000,
			want:      0xFFFFFFFF,
			wantN:     1,
			wantC:     1,
		},
		{
			name:      "ASR encoded zero means 32 for positive input",
			operation: 2,
			amount:    0,
			input:     0x7FFFFFFF,
			initialC:  1,
			want:      0,
			wantZ:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			operationName := []string{"LSL", "LSR", "ASR"}[tt.operation]
			effectiveAmount := tt.amount

			if tt.operation != 0 && effectiveAmount == 0 {
				effectiveAmount = 32
			}

			resultReason := fmt.Sprintf(
				"%s shifts %#08x by %d bits",
				operationName, tt.input, effectiveAmount,
			)

			switch tt.operation {
			case 0:
				resultReason += ", filling with zeros and discarding bits beyond bit 31"
			case 1:
				resultReason += ", filling the upper bits with zeros"
			case 2:
				resultReason += ", filling the upper bits with the original sign bit"
			}

			carryReason := ""
			if tt.operation == 0 {
				if tt.amount == 0 {
					carryReason = "LSL #0 preserves the original carry flag"
				} else {
					carryReason = fmt.Sprintf(
						"LSL #%d copies original bit %d into C",
						tt.amount, 32-tt.amount,
					)
				}
			} else {
				carryReason = fmt.Sprintf(
					"%s by %d copies original bit %d into C",
					operationName, effectiveAmount, effectiveAmount-1,
				)
			}

			for _, initialV := range []uint32{0, 1} {
				for _, rd := range []uint16{2, 1} {
					name := fmt.Sprintf("V=%d/Rd=%d", initialV, rd)

					t.Run(name, func(t *testing.T) {
						c := CPU{
							Registers: &Registers{},
						}

						const rs uint16 = 1
						const flagMask CPSR = 0xF0000000
						const otherBits CPSR = 0x0000003F

						c.Registers.Common[rs] = tt.input
						c.Registers.CPSR = otherBits

						// Start N and Z opposite to their expected values.
						c.Registers.SetCPSRBit(N, 1-tt.wantN)
						c.Registers.SetCPSRBit(Z, 1-tt.wantZ)
						c.Registers.SetCPSRBit(C, tt.initialC)
						c.Registers.SetCPSRBit(V, initialV)

						// Bits 15–13 = 000
						// Bits 12–11 = operation
						// Bits 10–6  = encoded shift amount
						// Bits 5–3   = Rs
						// Bits 2–0   = Rd
						opcode := tt.operation<<11 |
							tt.amount<<6 |
							rs<<3 |
							rd

						c.MoveShiftedRegister(opcode)

						if got := c.Registers.Common[rd]; got != tt.want {
							t.Errorf(
								"Result mismatch\nReason: %s\nExpected: %#08x\nGot:      %#08x",
								resultReason, tt.want, got,
							)
						}

						if rd != rs {
							if got := c.Registers.Common[rs]; got != tt.input {
								t.Errorf(
									"Source register changed\nReason: Rs must retain its original value when Rd differs from Rs\nExpected: %#08x\nGot:      %#08x",
									tt.input, got,
								)
							}
						}

						flagChecks := []struct {
							name   string
							index  CPSRIndex
							want   uint32
							reason string
						}{
							{
								name:  "N",
								index: N,
								want:  tt.wantN,
								reason: fmt.Sprintf(
									"N must equal bit 31 of the expected result %#08x",
									tt.want,
								),
							},
							{
								name:   "Z",
								index:  Z,
								want:   tt.wantZ,
								reason: "Z must be 1 when the 32-bit result is zero, otherwise 0",
							},
							{
								name:   "C",
								index:  C,
								want:   tt.wantC,
								reason: carryReason,
							},
							{
								name:   "V",
								index:  V,
								want:   initialV,
								reason: "MoveShiftedRegister must preserve the original V flag",
							},
						}

						for _, flag := range flagChecks {
							if got := c.Registers.GetCPSRBit(flag.index); got != flag.want {
								t.Errorf(
									"%s flag mismatch\nReason: %s\nExpected: %d\nGot:      %d",
									flag.name, flag.reason, flag.want, got,
								)
							}
						}

						if got := c.Registers.CPSR &^ flagMask; got != otherBits {
							t.Errorf(
								"CPSR mismatch\nReason: All CPSR bits outside NZCV must be preserved\nExpected: %#08x\nGot:      %#08x",
								otherBits, got,
							)
						}
					})
				}
			}
		})
	}
}

const (
	thumbTestN uint32 = 8
	thumbTestZ uint32 = 4
	thumbTestC uint32 = 2
	thumbTestV uint32 = 1

	thumbTestNZCV = thumbTestN | thumbTestZ | thumbTestC | thumbTestV
)

// checkThumbFlags checks selected flags and preservation of other CPSR bits.
func checkThumbFlags(
	t *testing.T,
	registers *Registers,
	want uint32,
	checkMask uint32,
	original CPSR,
	reason string,
) {
	t.Helper()

	flags := []struct {
		name  string
		index CPSRIndex
		mask  uint32
	}{
		{"N", N, thumbTestN},
		{"Z", Z, thumbTestZ},
		{"C", C, thumbTestC},
		{"V", V, thumbTestV},
	}

	for _, flag := range flags {
		if checkMask&flag.mask == 0 {
			continue
		}

		expected := uint32(0)
		if want&flag.mask != 0 {
			expected = 1
		}

		got := registers.GetCPSRBit(flag.index)
		if got != expected {
			t.Errorf(
				"%s flag mismatch\nReason: %s\nExpected: %d\nGot:      %d",
				flag.name, reason, expected, got,
			)
		}
	}

	const flagMask CPSR = 0xF0000000
	wantOther := original &^ flagMask
	gotOther := registers.CPSR &^ flagMask

	if gotOther != wantOther {
		t.Errorf(
			"CPSR mismatch\nReason: Bits outside NZCV must be preserved\nExpected: %#08x\nGot:      %#08x",
			wantOther, gotOther,
		)
	}
}

func setThumbTestFlags(registers *Registers, flags uint32) {
	// System mode with the Thumb bit set.
	registers.CPSR = CPSR(0x3F) | CPSR(flags)<<28
}

func checkThumbRegisters(
	t *testing.T,
	got CommonRegisters,
	want CommonRegisters,
	reason string,
) {
	t.Helper()

	for index := range want {
		if got[index] != want[index] {
			t.Errorf(
				"R%d mismatch\nReason: %s; only the destination may change when the instruction writes a result\nExpected: %#08x\nGot:      %#08x",
				index, reason, want[index], got[index],
			)
		}
	}
}

func TestAddSub(t *testing.T) {
	tests := []struct {
		name   string
		op     uint16
		a      uint32
		b      uint32 // Register value or 3-bit immediate
		want   uint32
		flags  uint32
		reason string
	}{
		{
			name:   "ADD register ordinary",
			op:     0,
			a:      5,
			b:      3,
			want:   8,
			reason: "5 + 3 = 8, with no carry or signed overflow",
		},
		{
			name:   "ADD register wraps to zero",
			op:     0,
			a:      0xFFFFFFFF,
			b:      1,
			want:   0,
			flags:  thumbTestZ | thumbTestC,
			reason: "Unsigned addition carries out; signed -1 + 1 fits",
		},
		{
			name:   "ADD register positive signed overflow",
			op:     0,
			a:      0x7FFFFFFF,
			b:      1,
			want:   0x80000000,
			flags:  thumbTestN | thumbTestV,
			reason: "Two nonnegative operands produce a negative signed result",
		},
		{
			name:   "ADD register negative signed overflow",
			op:     0,
			a:      0x80000000,
			b:      0x80000000,
			want:   0,
			flags:  thumbTestZ | thumbTestC | thumbTestV,
			reason: "Two negative operands overflow to zero and produce unsigned carry",
		},
		{
			name:   "ADD register negative result without overflow",
			op:     0,
			a:      0xFFFFFFFF,
			b:      0xFFFFFFFF,
			want:   0xFFFFFFFE,
			flags:  thumbTestN | thumbTestC,
			reason: "Signed -1 + -1 = -2 fits; unsigned addition carries out",
		},
		{
			name:   "SUB register ordinary",
			op:     1,
			a:      5,
			b:      3,
			want:   2,
			flags:  thumbTestC,
			reason: "5 - 3 = 2 with no unsigned borrow",
		},
		{
			name:   "SUB register equal operands",
			op:     1,
			a:      5,
			b:      5,
			want:   0,
			flags:  thumbTestZ | thumbTestC,
			reason: "Equal operands produce zero with no borrow",
		},
		{
			name:   "SUB register borrow",
			op:     1,
			a:      3,
			b:      5,
			want:   0xFFFFFFFE,
			flags:  thumbTestN,
			reason: "3 - 5 wraps to -2 and requires an unsigned borrow",
		},
		{
			name:   "SUB register negative minus positive overflow",
			op:     1,
			a:      0x80000000,
			b:      1,
			want:   0x7FFFFFFF,
			flags:  thumbTestC | thumbTestV,
			reason: "Signed minimum minus 1 overflows; unsigned operands need no borrow",
		},
		{
			name:   "SUB register positive minus negative overflow",
			op:     1,
			a:      0x7FFFFFFF,
			b:      0xFFFFFFFF,
			want:   0x80000000,
			flags:  thumbTestN | thumbTestV,
			reason: "Signed maximum minus -1 overflows; unsigned subtraction borrows",
		},
		{
			name:   "ADD immediate maximum immediate",
			op:     2,
			a:      5,
			b:      7,
			want:   12,
			reason: "The 3-bit immediate 7 is added directly",
		},
		{
			name:   "ADD immediate zero updates flags",
			op:     2,
			a:      0x80000000,
			b:      0,
			want:   0x80000000,
			flags:  thumbTestN,
			reason: "Adding zero preserves the value and clears C and V",
		},
		{
			name:   "ADD immediate carry",
			op:     2,
			a:      0xFFFFFFFF,
			b:      1,
			want:   0,
			flags:  thumbTestZ | thumbTestC,
			reason: "Adding 1 wraps to zero with unsigned carry",
		},
		{
			name:   "ADD immediate signed overflow",
			op:     2,
			a:      0x7FFFFFFF,
			b:      1,
			want:   0x80000000,
			flags:  thumbTestN | thumbTestV,
			reason: "Signed maximum plus 1 overflows",
		},
		{
			name:   "SUB immediate ordinary",
			op:     3,
			a:      10,
			b:      7,
			want:   3,
			flags:  thumbTestC,
			reason: "10 - 7 = 3 with no borrow",
		},
		{
			name:   "SUB immediate borrow",
			op:     3,
			a:      0,
			b:      1,
			want:   0xFFFFFFFF,
			flags:  thumbTestN,
			reason: "0 - 1 requires a borrow and produces signed -1",
		},
		{
			name:   "SUB immediate signed overflow",
			op:     3,
			a:      0x80000000,
			b:      1,
			want:   0x7FFFFFFF,
			flags:  thumbTestC | thumbTestV,
			reason: "Signed minimum minus 1 overflows",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// R0 is separate, R1 aliases the first operand,
			// and R2 aliases the second register operand.
			for _, rd := range []uint16{0, 1, 2} {
				for _, initialFlags := range []uint32{0, thumbTestNZCV} {
					c := CPU{Registers: &Registers{}}
					c.Registers.Common[0] = 0xDEADBEEF
					c.Registers.Common[1] = tt.a
					c.Registers.Common[2] = tt.b
					setThumbTestFlags(c.Registers, initialFlags)

					originalCPSR := c.Registers.CPSR
					wantRegisters := c.Registers.Common
					wantRegisters[rd] = tt.want

					const rs uint16 = 1
					operandField := uint16(2)
					if tt.op >= 2 {
						operandField = uint16(tt.b)
					}

					opcode := uint16(0x1800) |
						tt.op<<9 |
						operandField<<6 |
						rs<<3 |
						rd

					c.AddSub(opcode)

					t.Logf("Rd=%d, initial NZCV=%04b", rd, initialFlags)
					checkThumbRegisters(
						t, c.Registers.Common, wantRegisters, tt.reason,
					)
					checkThumbFlags(
						t, c.Registers, tt.flags, thumbTestNZCV,
						originalCPSR, tt.reason,
					)
				}
			}
		})
	}
}

func TestMovCmpAddSubImm(t *testing.T) {
	tests := []struct {
		name   string
		op     uint16
		a      uint32
		nn     uint16
		want   uint32
		flags  uint32
		reason string
	}{
		{
			name:   "MOV zero",
			op:     0,
			a:      0xFFFFFFFF,
			nn:     0,
			want:   0,
			flags:  thumbTestZ,
			reason: "MOV zero sets Z and clears N while preserving C and V",
		},
		{
			name:   "MOV full 8-bit immediate",
			op:     0,
			a:      0,
			nn:     255,
			want:   255,
			reason: "MOV uses all 8 immediate bits and preserves C and V",
		},
		{
			name:   "CMP equal",
			op:     1,
			a:      200,
			nn:     200,
			want:   200,
			flags:  thumbTestZ | thumbTestC,
			reason: "Equal values set Z and C; CMP preserves Rd",
		},
		{
			name:   "CMP greater",
			op:     1,
			a:      10,
			nn:     3,
			want:   10,
			flags:  thumbTestC,
			reason: "10 - 3 needs no borrow; CMP preserves Rd",
		},
		{
			name:   "CMP less",
			op:     1,
			a:      3,
			nn:     5,
			want:   3,
			flags:  thumbTestN,
			reason: "3 - 5 produces a negative temporary result and borrows",
		},
		{
			name:   "CMP signed overflow",
			op:     1,
			a:      0x80000000,
			nn:     1,
			want:   0x80000000,
			flags:  thumbTestC | thumbTestV,
			reason: "The temporary signed subtraction overflows; Rd is preserved",
		},
		{
			name:   "ADD full 8-bit immediate",
			op:     2,
			a:      1,
			nn:     255,
			want:   256,
			reason: "1 + 255 = 256 with no carry or overflow",
		},
		{
			name:   "ADD carry",
			op:     2,
			a:      0xFFFFFFFF,
			nn:     1,
			want:   0,
			flags:  thumbTestZ | thumbTestC,
			reason: "Unsigned addition wraps to zero; signed -1 + 1 fits",
		},
		{
			name:   "ADD signed overflow",
			op:     2,
			a:      0x7FFFFFFF,
			nn:     1,
			want:   0x80000000,
			flags:  thumbTestN | thumbTestV,
			reason: "Signed maximum plus 1 overflows",
		},
		{
			name:   "ADD negative becomes zero without signed overflow",
			op:     2,
			a:      0xFFFFFFFF,
			nn:     1,
			want:   0,
			flags:  thumbTestZ | thumbTestC,
			reason: "Different-sign addition cannot produce signed overflow",
		},
		{
			name:   "SUB full 8-bit immediate",
			op:     3,
			a:      256,
			nn:     255,
			want:   1,
			flags:  thumbTestC,
			reason: "256 - 255 = 1 with no borrow",
		},
		{
			name:   "SUB equal",
			op:     3,
			a:      255,
			nn:     255,
			want:   0,
			flags:  thumbTestZ | thumbTestC,
			reason: "Equal operands produce zero with no borrow",
		},
		{
			name:   "SUB borrow without signed overflow",
			op:     3,
			a:      0,
			nn:     1,
			want:   0xFFFFFFFF,
			flags:  thumbTestN,
			reason: "0 - 1 = -1 fits as signed but needs an unsigned borrow",
		},
		{
			name:   "SUB signed overflow",
			op:     3,
			a:      0x80000000,
			nn:     1,
			want:   0x7FFFFFFF,
			flags:  thumbTestC | thumbTestV,
			reason: "Signed minimum minus 1 overflows",
		},
		{
			name:   "SUB negative without overflow",
			op:     3,
			a:      0xFFFFFFFF,
			nn:     1,
			want:   0xFFFFFFFE,
			flags:  thumbTestN | thumbTestC,
			reason: "Signed -1 - 1 = -2 fits, and unsigned subtraction needs no borrow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, rd := range []uint16{0, 3, 7} {
				for _, initialFlags := range []uint32{0, thumbTestNZCV} {
					c := CPU{Registers: &Registers{}}
					c.Registers.Common[rd] = tt.a
					setThumbTestFlags(c.Registers, initialFlags)

					originalCPSR := c.Registers.CPSR
					wantRegisters := c.Registers.Common
					wantRegisters[rd] = tt.want

					wantFlags := tt.flags
					if tt.op == 0 {
						wantFlags |= initialFlags & (thumbTestC | thumbTestV)
					}

					opcode := uint16(0x2000) |
						tt.op<<11 |
						rd<<8 |
						tt.nn

					c.MovCmpAddSubImm(opcode)

					t.Logf("Rd=%d, initial NZCV=%04b", rd, initialFlags)
					checkThumbRegisters(
						t, c.Registers.Common, wantRegisters, tt.reason,
					)
					checkThumbFlags(
						t, c.Registers, wantFlags, thumbTestNZCV,
						originalCPSR, tt.reason,
					)
				}
			}
		})
	}
}

func TestALUOperations(t *testing.T) {
	tests := []struct {
		name    string
		op      uint16
		a       uint32
		b       uint32
		initial uint32
		want    uint32
		flags   uint32
		reason  string
	}{
		{
			name:    "AND bitwise result",
			op:      0x0,
			a:       0xF0,
			b:       0x3C,
			initial: thumbTestNZCV,
			want:    0x30,
			flags:   thumbTestC | thumbTestV,
			reason:  "Opcode 0 is AND; 0xF0 AND 0x3C = 0x30; C and V are preserved",
		},
		{
			name:    "AND zero",
			op:      0x0,
			a:       0xF0,
			b:       0x0F,
			initial: thumbTestNZCV,
			want:    0,
			flags:   thumbTestZ | thumbTestC | thumbTestV,
			reason:  "Disjoint bit patterns produce zero",
		},
		{
			name:    "EOR identical operands",
			op:      0x1,
			a:       0xAAAAAAAA,
			b:       0xAAAAAAAA,
			initial: thumbTestNZCV,
			want:    0,
			flags:   thumbTestZ | thumbTestC | thumbTestV,
			reason:  "A value XOR itself is zero; C and V are preserved",
		},
		{
			name:   "EOR negative result",
			op:     0x1,
			a:      0x80000000,
			b:      1,
			want:   0x80000001,
			flags:  thumbTestN,
			reason: "XOR sets the differing bits and the result has bit 31 set",
		},
		{
			name:    "LSL zero preserves carry",
			op:      0x2,
			a:       0x80000001,
			b:       0,
			initial: thumbTestC | thumbTestV,
			want:    0x80000001,
			flags:   thumbTestN | thumbTestC | thumbTestV,
			reason:  "A register shift count of zero preserves the value and C",
		},
		{
			name:    "LSL one",
			op:      0x2,
			a:       0x80000001,
			b:       1,
			initial: thumbTestV,
			want:    2,
			flags:   thumbTestC | thumbTestV,
			reason:  "LSL 1 shifts original bit 31 into C and preserves V",
		},
		{
			name:    "LSL 32",
			op:      0x2,
			a:       1,
			b:       32,
			initial: thumbTestV,
			want:    0,
			flags:   thumbTestZ | thumbTestC | thumbTestV,
			reason:  "LSL 32 produces zero and copies original bit 0 into C",
		},
		{
			name:    "LSL 33",
			op:      0x2,
			a:       0xFFFFFFFF,
			b:       33,
			initial: thumbTestC | thumbTestV,
			want:    0,
			flags:   thumbTestZ | thumbTestV,
			reason:  "LSL greater than 32 produces zero and clears C",
		},
		{
			name:   "LSL uses low eight count bits",
			op:     0x2,
			a:      1,
			b:      0x101,
			want:   2,
			reason: "0x101 AND 0xFF gives a shift count of 1",
		},
		{
			name:    "LSR zero preserves carry",
			op:      0x3,
			a:       0x80000000,
			b:       0,
			initial: thumbTestC | thumbTestV,
			want:    0x80000000,
			flags:   thumbTestN | thumbTestC | thumbTestV,
			reason:  "Register LSR count zero preserves the value and C",
		},
		{
			name:    "LSR one",
			op:      0x3,
			a:       0x80000001,
			b:       1,
			initial: thumbTestV,
			want:    0x40000000,
			flags:   thumbTestC | thumbTestV,
			reason:  "LSR fills with zero and copies original bit 0 into C",
		},
		{
			name:   "LSR 32",
			op:     0x3,
			a:      0x80000000,
			b:      32,
			want:   0,
			flags:  thumbTestZ | thumbTestC,
			reason: "LSR 32 produces zero and copies original bit 31 into C",
		},
		{
			name:    "LSR 33",
			op:      0x3,
			a:       0xFFFFFFFF,
			b:       33,
			initial: thumbTestC | thumbTestV,
			want:    0,
			flags:   thumbTestZ | thumbTestV,
			reason:  "LSR greater than 32 produces zero and clears C",
		},
		{
			name:    "ASR zero preserves carry",
			op:      0x4,
			a:       0x80000000,
			b:       0,
			initial: thumbTestC | thumbTestV,
			want:    0x80000000,
			flags:   thumbTestN | thumbTestC | thumbTestV,
			reason:  "Register ASR count zero preserves the value and C",
		},
		{
			name:    "ASR negative by one",
			op:      0x4,
			a:       0x80000001,
			b:       1,
			initial: thumbTestV,
			want:    0xC0000000,
			flags:   thumbTestN | thumbTestC | thumbTestV,
			reason:  "ASR replicates the original sign bit and shifts bit 0 into C",
		},
		{
			name:   "ASR negative by 32",
			op:     0x4,
			a:      0x80000000,
			b:      32,
			want:   0xFFFFFFFF,
			flags:  thumbTestN | thumbTestC,
			reason: "ASR 32 fills the result and C with the original sign bit",
		},
		{
			name:    "ASR negative by 255",
			op:      0x4,
			a:       0x80000000,
			b:       255,
			initial: thumbTestV,
			want:    0xFFFFFFFF,
			flags:   thumbTestN | thumbTestC | thumbTestV,
			reason:  "ASR counts above 32 retain sign-filled result and sign-bit carry",
		},
		{
			name:    "ASR positive by 33",
			op:      0x4,
			a:       0x7FFFFFFF,
			b:       33,
			initial: thumbTestC | thumbTestV,
			want:    0,
			flags:   thumbTestZ | thumbTestV,
			reason:  "Positive ASR by at least 32 produces zero and clears C",
		},
		{
			name:    "ADC includes set carry",
			op:      0x5,
			a:       5,
			b:       3,
			initial: thumbTestC | thumbTestV,
			want:    9,
			reason:  "ADC computes 5 + 3 + old C = 9 and recalculates C and V",
		},
		{
			name:    "ADC carry out",
			op:      0x5,
			a:       0xFFFFFFFF,
			b:       0,
			initial: thumbTestC,
			want:    0,
			flags:   thumbTestZ | thumbTestC,
			reason:  "0xFFFFFFFF + 0 + 1 wraps to zero with carry",
		},
		{
			name:    "ADC signed overflow from carry input",
			op:      0x5,
			a:       0x7FFFFFFF,
			b:       0,
			initial: thumbTestC,
			want:    0x80000000,
			flags:   thumbTestN | thumbTestV,
			reason:  "The incoming carry pushes signed maximum beyond its limit",
		},
		{
			name:   "ADC clear carry",
			op:     0x5,
			a:      5,
			b:      3,
			want:   8,
			reason: "ADC with old C clear computes 5 + 3",
		},
		{
			name:    "SBC set carry subtracts no extra borrow",
			op:      0x6,
			a:       5,
			b:       3,
			initial: thumbTestC,
			want:    2,
			flags:   thumbTestC,
			reason:  "SBC computes a - b - (1 - old C)",
		},
		{
			name:   "SBC clear carry subtracts extra one",
			op:     0x6,
			a:      5,
			b:      3,
			want:   1,
			flags:  thumbTestC,
			reason: "With old C clear, SBC computes 5 - 3 - 1",
		},
		{
			name:   "SBC equal operands with incoming borrow",
			op:     0x6,
			a:      5,
			b:      5,
			want:   0xFFFFFFFF,
			flags:  thumbTestN,
			reason: "5 - 5 - 1 requires an unsigned borrow",
		},
		{
			name:   "SBC operand plus borrow exceeds uint32",
			op:     0x6,
			a:      0xFFFFFFFF,
			b:      0xFFFFFFFF,
			want:   0xFFFFFFFF,
			flags:  thumbTestN,
			reason: "The full subtrahend is 0x100000000, so C must clear",
		},
		{
			name:   "SBC signed overflow",
			op:     0x6,
			a:      0x80000000,
			b:      0,
			want:   0x7FFFFFFF,
			flags:  thumbTestC | thumbTestV,
			reason: "Signed minimum minus the incoming borrow overflows",
		},
		{
			name:    "ROR zero preserves carry",
			op:      0x7,
			a:       1,
			b:       0,
			initial: thumbTestC | thumbTestV,
			want:    1,
			flags:   thumbTestC | thumbTestV,
			reason:  "ROR count zero preserves the value, C, and V",
		},
		{
			name:    "ROR one wraps bit zero",
			op:      0x7,
			a:       1,
			b:       1,
			initial: thumbTestV,
			want:    0x80000000,
			flags:   thumbTestN | thumbTestC | thumbTestV,
			reason:  "Bit 0 rotates into bit 31 and C",
		},
		{
			name:   "ROR 32 updates carry",
			op:     0x7,
			a:      0x80000001,
			b:      32,
			want:   0x80000001,
			flags:  thumbTestN | thumbTestC,
			reason: "A nonzero full rotation copies result bit 31 into C",
		},
		{
			name:   "ROR 33",
			op:     0x7,
			a:      1,
			b:      33,
			want:   0x80000000,
			flags:  thumbTestN | thumbTestC,
			reason: "Rotating by 33 has the same result as rotating by 1",
		},
		{
			name:    "ROR count low byte zero",
			op:      0x7,
			a:       1,
			b:       0x100,
			initial: thumbTestC,
			want:    1,
			flags:   thumbTestC,
			reason:  "Only the low 8 count bits are used, making this a zero rotation",
		},
		{
			name:    "TST zero preserves destination",
			op:      0x8,
			a:       0xF0,
			b:       0x0F,
			initial: thumbTestC | thumbTestV,
			want:    0xF0,
			flags:   thumbTestZ | thumbTestC | thumbTestV,
			reason:  "Temporary AND is zero; Rd, C, and V are preserved",
		},
		{
			name:   "TST negative temporary result",
			op:     0x8,
			a:      0x80000001,
			b:      0x80000000,
			want:   0x80000001,
			flags:  thumbTestN,
			reason: "Temporary AND has bit 31 set; Rd is preserved",
		},
		{
			name:    "NEG zero",
			op:      0x9,
			a:       123,
			b:       0,
			initial: thumbTestV,
			want:    0,
			flags:   thumbTestZ | thumbTestC,
			reason:  "0 - 0 is zero with no borrow or signed overflow",
		},
		{
			name:   "NEG positive",
			op:     0x9,
			a:      123,
			b:      1,
			want:   0xFFFFFFFF,
			flags:  thumbTestN,
			reason: "0 - 1 produces -1 and requires a borrow",
		},
		{
			name:   "NEG signed minimum",
			op:     0x9,
			a:      0x80000000,
			b:      0x80000000,
			want:   0x80000000,
			flags:  thumbTestN | thumbTestV,
			reason: "Negating signed minimum overflows; the first operand is zero",
		},
		{
			name:   "CMP equal",
			op:     0xA,
			a:      5,
			b:      5,
			want:   5,
			flags:  thumbTestZ | thumbTestC,
			reason: "Temporary subtraction is zero with no borrow; Rd is preserved",
		},
		{
			name:   "CMP borrow",
			op:     0xA,
			a:      3,
			b:      5,
			want:   3,
			flags:  thumbTestN,
			reason: "Temporary 3 - 5 is negative and borrows",
		},
		{
			name:   "CMP signed overflow",
			op:     0xA,
			a:      0x80000000,
			b:      1,
			want:   0x80000000,
			flags:  thumbTestC | thumbTestV,
			reason: "Signed minimum minus 1 overflows; Rd is preserved",
		},
		{
			name:   "CMN carry",
			op:     0xB,
			a:      0xFFFFFFFF,
			b:      1,
			want:   0xFFFFFFFF,
			flags:  thumbTestZ | thumbTestC,
			reason: "Temporary addition wraps to zero with carry; Rd is preserved",
		},
		{
			name:   "CMN signed overflow",
			op:     0xB,
			a:      0x7FFFFFFF,
			b:      1,
			want:   0x7FFFFFFF,
			flags:  thumbTestN | thumbTestV,
			reason: "Temporary addition overflows signed range; Rd is preserved",
		},
		{
			name:    "ORR",
			op:      0xC,
			a:       0x80000000,
			b:       1,
			initial: thumbTestC | thumbTestV,
			want:    0x80000001,
			flags:   thumbTestN | thumbTestC | thumbTestV,
			reason:  "OR combines set bits and preserves C and V",
		},
		{
			name:    "MUL ordinary",
			op:      0xD,
			a:       3,
			b:       7,
			initial: thumbTestV,
			want:    21,
			flags:   thumbTestV,
			reason:  "3 * 7 = 21; MUL updates N and Z and preserves V",
		},
		{
			name:    "MUL wraps to zero",
			op:      0xD,
			a:       0x80000000,
			b:       2,
			initial: thumbTestV,
			want:    0,
			flags:   thumbTestZ | thumbTestV,
			reason:  "The low 32 product bits are zero; V is preserved",
		},
		{
			name:   "MUL negative result",
			op:     0xD,
			a:      0xFFFFFFFF,
			b:      2,
			want:   0xFFFFFFFE,
			flags:  thumbTestN,
			reason: "The low 32 product bits have bit 31 set",
		},
		{
			name:    "BIC",
			op:      0xE,
			a:       0xFF,
			b:       0x0F,
			initial: thumbTestC | thumbTestV,
			want:    0xF0,
			flags:   thumbTestC | thumbTestV,
			reason:  "BIC clears bits selected by Rs and preserves C and V",
		},
		{
			name:    "MVN zero input",
			op:      0xF,
			a:       123,
			b:       0,
			initial: thumbTestC | thumbTestV,
			want:    0xFFFFFFFF,
			flags:   thumbTestN | thumbTestC | thumbTestV,
			reason:  "Inverting zero sets every bit; C and V are preserved",
		},
		{
			name:   "MVN all bits set",
			op:     0xF,
			a:      123,
			b:      0xFFFFFFFF,
			want:   0,
			flags:  thumbTestZ,
			reason: "Inverting all set bits produces zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CPU{Registers: &Registers{}}

			const rd uint16 = 3
			const rs uint16 = 6

			c.Registers.Common[rd] = tt.a
			c.Registers.Common[rs] = tt.b
			setThumbTestFlags(c.Registers, tt.initial)

			originalCPSR := c.Registers.CPSR
			wantRegisters := c.Registers.Common
			wantRegisters[rd] = tt.want

			opcode := uint16(0x4000) | tt.op<<6 | rs<<3 | rd

			// Recover within each subtest so a bad shift count
			// does not prevent the remaining cases from running.
			defer func() {
				if recovered := recover(); recovered != nil {
					t.Errorf(
						"Instruction panicked\nReason: %s\nExpected: Successful execution\nGot:      %v",
						tt.reason, recovered,
					)
				}
			}()

			c.ALUOperations(opcode)

			checkMask := uint32(thumbTestNZCV)
			if tt.op == 0xD {
				// ARM7TDMI MUL carry is unspecified.
				checkMask &^= thumbTestC
			}

			checkThumbRegisters(
				t, c.Registers.Common, wantRegisters, tt.reason,
			)
			checkThumbFlags(
				t, c.Registers, tt.flags, checkMask,
				originalCPSR, tt.reason,
			)
		})
	}
}

func TestHiRegisterOperations(t *testing.T) {
	const (
		baseCPSR CPSR = 0x0000003F // System mode, Thumb state
		flagMask CPSR = 0xF0000000
		allFlags CPSR = flagMask
	)

	tests := []struct {
		name        string
		op          uint16
		rd          uint16
		rs          uint16
		destination uint32
		source      uint32
		pc          uint32
		initial     CPSR
		wantValue   uint32
		wantPC      uint32
		wantCPSR    CPSR
		reason      string
	}{
		{
			name:        "ADD low source to high destination",
			op:          0,
			rd:          8,
			rs:          1,
			destination: 5,
			source:      3,
			pc:          0x1004,
			initial:     allFlags,
			wantValue:   8,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | allFlags,
			reason:      "R8 receives 5 + 3; high-register ADD preserves NZCV",
		},
		{
			name:        "ADD high source to low destination",
			op:          0,
			rd:          2,
			rs:          9,
			destination: 10,
			source:      20,
			pc:          0x1004,
			wantValue:   30,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR,
			reason:      "The source extension bit selects R9",
		},
		{
			name:        "ADD wraps without changing flags",
			op:          0,
			rd:          10,
			rs:          11,
			destination: 0xFFFFFFFF,
			source:      1,
			pc:          0x1004,
			initial:     CPSR(1) << V,
			wantValue:   0,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | CPSR(1)<<V,
			reason:      "The result wraps to zero, but high-register ADD preserves NZCV",
		},
		{
			name:        "ADD same register",
			op:          0,
			rd:          8,
			rs:          8,
			destination: 7,
			source:      7,
			pc:          0x1004,
			initial:     allFlags,
			wantValue:   14,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | allFlags,
			reason:      "Both operands read the original R8 value",
		},
		{
			name:        "ADD reads PC operand",
			op:          0,
			rd:          8,
			rs:          15,
			destination: 2,
			source:      0x1006,
			pc:          0x1006,
			wantValue:   0x1008,
			wantPC:      0x1006,
			wantCPSR:    baseCPSR,
			reason:      "An instruction at 0x1002 reads PC as 0x1006",
		},
		{
			name:        "ADD writes PC and stays Thumb",
			op:          0,
			rd:          15,
			rs:          8,
			destination: 0x1004,
			source:      3,
			pc:          0x1004,
			initial:     allFlags,
			wantValue:   0x1006,
			wantPC:      0x1006,
			wantCPSR:    baseCPSR | allFlags,
			reason:      "ADD PC clears target bit 0 and preserves Thumb state and NZCV",
		},
		{
			name:        "CMP equal",
			op:          1,
			rd:          8,
			rs:          1,
			destination: 5,
			source:      5,
			pc:          0x1004,
			initial:     allFlags,
			wantValue:   5,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | CPSR(1)<<Z | CPSR(1)<<C,
			reason:      "Equal operands set Z and C; Rd remains unchanged",
		},
		{
			name:        "CMP unsigned borrow",
			op:          1,
			rd:          1,
			rs:          8,
			destination: 3,
			source:      5,
			pc:          0x1004,
			initial:     allFlags,
			wantValue:   3,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | CPSR(1)<<N,
			reason:      "3 - 5 produces -2 with a borrow and no signed overflow",
		},
		{
			name:        "CMP greater",
			op:          1,
			rd:          9,
			rs:          10,
			destination: 10,
			source:      3,
			pc:          0x1004,
			initial:     allFlags,
			wantValue:   10,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | CPSR(1)<<C,
			reason:      "10 - 3 is positive with no unsigned borrow",
		},
		{
			name:        "CMP negative minus positive overflow",
			op:          1,
			rd:          8,
			rs:          1,
			destination: 0x80000000,
			source:      1,
			pc:          0x1004,
			wantValue:   0x80000000,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | CPSR(1)<<C | CPSR(1)<<V,
			reason:      "Signed minimum minus 1 overflows; unsigned subtraction has no borrow",
		},
		{
			name:        "CMP positive minus negative overflow",
			op:          1,
			rd:          8,
			rs:          9,
			destination: 0x7FFFFFFF,
			source:      0xFFFFFFFF,
			pc:          0x1004,
			wantValue:   0x7FFFFFFF,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | CPSR(1)<<N | CPSR(1)<<V,
			reason:      "Signed maximum minus -1 overflows and requires an unsigned borrow",
		},
		{
			name:        "CMP reads PC operand",
			op:          1,
			rd:          8,
			rs:          15,
			destination: 0x1006,
			source:      0x1006,
			pc:          0x1006,
			wantValue:   0x1006,
			wantPC:      0x1006,
			wantCPSR:    baseCPSR | CPSR(1)<<Z | CPSR(1)<<C,
			reason:      "PC already contains instruction address plus 4, making the operands equal",
		},
		{
			name:        "MOV low source to high destination",
			op:          2,
			rd:          8,
			rs:          1,
			destination: 0xDEADBEEF,
			source:      0,
			pc:          0x1004,
			initial:     allFlags,
			wantValue:   0,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | allFlags,
			reason:      "MOV copies zero and preserves all flags",
		},
		{
			name:        "MOV high source to low destination",
			op:          2,
			rd:          2,
			rs:          12,
			destination: 0,
			source:      0x87654321,
			pc:          0x1004,
			wantValue:   0x87654321,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR,
			reason:      "The source extension bit selects R12; MOV preserves NZCV",
		},
		{
			name:        "MOV same register acts as NOP",
			op:          2,
			rd:          8,
			rs:          8,
			destination: 0x12345678,
			source:      0x12345678,
			pc:          0x1004,
			initial:     allFlags,
			wantValue:   0x12345678,
			wantPC:      0x1004,
			wantCPSR:    baseCPSR | allFlags,
			reason:      "MOV R8,R8 preserves the register and CPSR",
		},
		{
			name:        "MOV reads PC",
			op:          2,
			rd:          8,
			rs:          15,
			destination: 0,
			source:      0x1006,
			pc:          0x1006,
			initial:     allFlags,
			wantValue:   0x1006,
			wantPC:      0x1006,
			wantCPSR:    baseCPSR | allFlags,
			reason:      "MOV reads PC as instruction address plus 4",
		},
		{
			name:        "MOV writes PC and stays Thumb",
			op:          2,
			rd:          15,
			rs:          8,
			destination: 0x1004,
			source:      0x08000103,
			pc:          0x1004,
			initial:     allFlags,
			wantValue:   0x08000102,
			wantPC:      0x08000102,
			wantCPSR:    baseCPSR | allFlags,
			reason:      "MOV PC clears target bit 0 and preserves Thumb state",
		},
		{
			name:        "MOV PC to PC branches",
			op:          2,
			rd:          15,
			rs:          15,
			destination: 0x1006,
			source:      0x1006,
			pc:          0x1006,
			initial:     allFlags,
			wantValue:   0x1006,
			wantPC:      0x1006,
			wantCPSR:    baseCPSR | allFlags,
			reason:      "MOV PC,PC targets instruction address plus 4 and stays Thumb",
		},
		{
			name:     "BX low register selects Thumb",
			op:       3,
			rs:       1,
			source:   0x08000103,
			pc:       0x1004,
			initial:  allFlags,
			wantPC:   0x08000102,
			wantCPSR: baseCPSR | allFlags,
			reason:   "Target bit 0 selects Thumb and is cleared from the branch address",
		},
		{
			name:     "BX high register selects ARM",
			op:       3,
			rs:       10,
			source:   0x08000100,
			pc:       0x1004,
			initial:  allFlags,
			wantPC:   0x08000100,
			wantCPSR: (baseCPSR &^ (CPSR(1) << THUMB_CPSR)) | allFlags,
			reason:   "An even target selects ARM while preserving NZCV",
		},
		{
			name:     "BX ARM target clears bit one",
			op:       3,
			rs:       8,
			source:   0x08000102,
			pc:       0x1004,
			wantPC:   0x08000100,
			wantCPSR: baseCPSR &^ (CPSR(1) << THUMB_CPSR),
			reason:   "ARM branch targets are aligned to four bytes",
		},
		{
			name:     "BX LR selects Thumb",
			op:       3,
			rs:       14,
			source:   0x08000201,
			pc:       0x1004,
			wantPC:   0x08000200,
			wantCPSR: baseCPSR,
			reason:   "An odd return address in LR selects Thumb",
		},
		{
			name:     "BX PC switches to ARM",
			op:       3,
			rs:       15,
			source:   0x1006,
			pc:       0x1006,
			initial:  allFlags,
			wantPC:   0x1004,
			wantCPSR: (baseCPSR &^ (CPSR(1) << THUMB_CPSR)) | allFlags,
			reason:   "At instruction address 0x1002, BX PC reads 0x1006 and aligns it to 0x1004",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := CPU{
				Registers: &Registers{},
			}

			// Distinct values help detect writes to the wrong register.
			for index := range c.Registers.Common {
				c.Registers.Common[index] = 0xA0000000 + uint32(index)
			}

			if tt.op != 3 {
				c.Registers.Common[tt.rd] = tt.destination
			}
			c.Registers.Common[tt.rs] = tt.source
			c.Registers.Common[PC] = tt.pc
			c.Registers.CPSR = baseCPSR | tt.initial

			wantRegisters := c.Registers.Common

			switch tt.op {
			case 0, 2:
				wantRegisters[tt.rd] = tt.wantValue
			}
			wantRegisters[PC] = tt.wantPC

			// Bits 9–8 hold the operation.
			// Bit 7 extends Rd, or selects BX/BLX for operation 3.
			// Bit 6 extends Rs.
			opcode := uint16(0x4400) |
				tt.op<<8 |
				(tt.rs&8)<<3 |
				(tt.rs&7)<<3

			if tt.op != 3 {
				opcode |= (tt.rd&8)<<4 | (tt.rd & 7)
			}
			// For BX, bit 7 and bits 2–0 remain zero.

			c.HiRegisterOperations(opcode)

			for index, expected := range wantRegisters {
				got := c.Registers.Common[index]
				if got != expected {
					t.Errorf(
						"R%d mismatch (opcode=%#04x)\nReason: %s; registers outside the instruction's write targets must be preserved\nExpected: %#08x\nGot:      %#08x",
						index, opcode, tt.reason, expected, got,
					)
				}
			}

			flags := []struct {
				name  string
				index CPSRIndex
			}{
				{"N", N},
				{"Z", Z},
				{"C", C},
				{"V", V},
				{"T", THUMB_CPSR},
			}

			for _, flag := range flags {
				expected := uint32(tt.wantCPSR>>flag.index) & 1
				got := c.Registers.GetCPSRBit(flag.index)

				if got != expected {
					t.Errorf(
						"%s flag mismatch\nReason: %s\nExpected: %d\nGot:      %d",
						flag.name, tt.reason, expected, got,
					)
				}
			}

			const checkedBits CPSR = flagMask | CPSR(1)<<THUMB_CPSR
			expectedOther := tt.wantCPSR &^ checkedBits
			gotOther := c.Registers.CPSR &^ checkedBits

			if gotOther != expectedOther {
				t.Errorf(
					"CPSR mismatch\nReason: Bits outside NZCV and T must be preserved\nExpected: %#08x\nGot:      %#08x",
					expectedOther, gotOther,
				)
			}
		})
	}
}
