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
