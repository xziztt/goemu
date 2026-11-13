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
