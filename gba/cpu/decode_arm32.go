package cpu

// apply bitmask on the opcode to see if it matches the pattern for a type.
func isMatch(opcode uint32, mask uint32, value uint32) bool {
	return (opcode & mask) == value
}

// Branch (B) is supposed to jump to a subroutine. Branch with Link is meant to be used to call to a subroutine, return address is then saved in R14.
//   Bit    Expl.
//   31-28  Condition (must be 1111b for BLX)
//   27-25  Must be "101" for this instruction
//   24     Opcode (0-1) (or Halfword Offset for BLX)
//           0: B{cond} label    ;branch            PC=PC+8+nn*4
//           1: BL{cond} label   ;branch/link       PC=PC+8+nn*4, LR=PC+4
//           H: BLX label ;ARM9  ;branch/link/thumb PC=PC+8+nn*4+H*2, LR=PC+4, T=1
//   23-0   nn - Signed Offset, step 4      (-32M..+32M in steps of 4)

func isBranchInstruction(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1110_0000_0000_0000_0000_0000_0000, 0b0000_1010_0000_0000_0000_0000_0000_0000)
}

// Opcode Format

//   Bit    Expl.
//   31-28  Condition
//   27-8   Must be "0001.0010.1111.1111.1111" for this instruction
//   7-4    Opcode
//           0001b: BX{cond}  Rn    ;PC=Rn, T=Rn.0  (ARMv4T and ARMv5 and up)
//           0011b: BLX{cond} Rn    ;PC=Rn, T=Rn.0, LR=PC+4    (ARMv5 and up)
//   3-0    Rn - Operand Register  (R0-R14)

// Switching to THUMB Mode: Set Bit 0 of the value in Rn to 1, program continues then at Rn-1 in THUMB mode.
// Results in undefined behaviour if using R15 (PC+8 itself) as operand.
// Execution Time: 2S + 1N
// Return: No flags affected.

func isBranchAndExchangeInstruction(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1111_1111_1111_1111_1111_1111_1111, 0b0000_0001_0010_1111_1111_1111_1111_1111)
}

// Opcode Format

//   Bit    Expl.
//   31-28  Condition
//   27-25  Must be 000b for this instruction
//   24-21  Opcode
//           0000b: MUL{cond}{S}   Rd,Rm,Rs        ;multiply   Rd = Rm*Rs
//           0001b: MLA{cond}{S}   Rd,Rm,Rs,Rn     ;mul.& accumulate Rd = Rm*Rs+Rn
//           0100b: UMULL{cond}{S} RdLo,RdHi,Rm,Rs ;multiply   RdHiLo=Rm*Rs
//           0101b: UMLAL{cond}{S} RdLo,RdHi,Rm,Rs ;mul.& acc. RdHiLo=Rm*Rs+RdHiLo
//           0110b: SMULL{cond}{S} RdLo,RdHi,Rm,Rs ;sign.mul.  RdHiLo=Rm*Rs
//           0111b: SMLAL{cond}{S} RdLo,RdHi,Rm,Rs ;sign.m&a.  RdHiLo=Rm*Rs+RdHiLo
//           1000b: SMLAxy{cond}   Rd,Rm,Rs,Rn     ;Rd=HalfRm*HalfRs+Rn
//           1001b: SMLAWy{cond}   Rd,Rm,Rs,Rn     ;Rd=(Rm*HalfRs)/10000h+Rn
//           1001b: SMULWy{cond}   Rd,Rm,Rs        ;Rd=(Rm*HalfRs)/10000h
//           1010b: SMLALxy{cond}  RdLo,RdHi,Rm,Rs ;RdHiLo=RdHiLo+HalfRm*HalfRs
//           1011b: SMULxy{cond}   Rd,Rm,Rs        ;Rd=HalfRm*HalfRs
//   20     S - Set Condition Codes (0=No, 1=Yes) (Must be 0 for Halfword mul)
//   19-16  Rd (or RdHi) - Destination Register (R0-R14)
//   15-12  Rn (or RdLo) - Accumulate Register  (R0-R14) (Set to 0000b if unused)
//   11-8   Rs - Operand Register               (R0-R14)
//   For Non-Halfword Multiplies
//     7-4  Must be 1001b for these instructions
//   For Halfword Multiplies
//     7    Must be 1 for these instructions
//     6    y - Rs Top/Bottom flag (0=B=Lower 16bit, 1=T=Upper 16bit)
//     5    x - Rm Top/Bottom flag (as above), or 0 for SMLAW, or 1 for SMULW
//     4    Must be 0 for these instructions
//   3-0    Rm - Operand Register               (R0-R14)

func isMultiply(opcode uint32) bool {
	// ideally we can check for 27-25, 7-5 and bit 23 (MUL/MLA vs Extended MUL)
	// In order to check if the opcode is mul related, we can use a single mask here
	return isMatch(opcode, 0b0000_1110_0000_0000_0000_0000_1111_0000, 0b0000_0000_0000_0000_0000_0000_1001_0000)
}

// Opcode Format

//   Bit    Expl.
//   31-28  Condition (Must be 1111b for PLD)
//   27-26  Must be 01b for this instruction
//   25     I - Immediate Offset Flag (0=Immediate, 1=Shifted Register)
//   24     P - Pre/Post (0=post; add offset after transfer, 1=pre; before trans.)
//   23     U - Up/Down Bit (0=down; subtract offset from base, 1=up; add to base)
//   22     B - Byte/Word bit (0=transfer word quantity, 1=transfer byte quantity)
//   When above Bit 24 P=0 (Post-indexing, write-back is ALWAYS enabled):
//     21     T - Memory Management (0=Normal, 1=Force non-privileged access)
//   When above Bit 24 P=1 (Pre-indexing, write-back is optional):
//     21     W - Write-back bit (0=no write-back, 1=write address into base)
//   20     L - Load/Store bit (0=Store to memory, 1=Load from memory)
//           0: STR{cond}{B}{T} Rd,<Address>   ;[Rn+/-<offset>]=Rd
//           1: LDR{cond}{B}{T} Rd,<Address>   ;Rd=[Rn+/-<offset>]
//          (1: PLD <Address> ;Prepare Cache for Load, see notes below)
//           Whereas, B=Byte, T=Force User Mode (only for POST-Indexing)
//   19-16  Rn - Base register               (R0..R15) (including R15=PC+8)
//   15-12  Rd - Source/Destination Register (R0..R15) (including R15=PC+12)
//   When above I=0 (Immediate as Offset)
//     11-0   Unsigned 12bit Immediate Offset (0-4095, steps of 1)
//   When above I=1 (Register shifted by Immediate as Offset)
//     11-7   Is - Shift amount      (1-31, 0=Special/See below)
//     6-5    Shift Type             (0=LSL, 1=LSR, 2=ASR, 3=ROR)
//     4      Must be 0 (Reserved, see ARM.17, The Undefined Instruction)
//     3-0    Rm - Offset Register   (R0..R14) (not including PC=R15)

// Checking for bits 27-26 only
func isSingleDataTransfer(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1100_0000_0000_0000_0000_0000_0000, 0b0000_0100_0000_0000_0000_0000_0000_0000)
}

// Opcode Format
// These instructions occupy an unused area (TEQ,TST,CMP,CMN with S=0) of Data Processing opcodes (ARM.5).

//   Bit    Expl.
//   31-28  Condition
//   27-26  Must be 00b for this instruction
//   25     I - Immediate Operand Flag  (0=Register, 1=Immediate) (Zero for MRS)
//   24-23  Must be 10b for this instruction
//   22     Psr - Source/Destination PSR  (0=CPSR, 1=SPSR_<current mode>)
//   21     Opcode
//            0: MRS{cond} Rd,Psr          ;Rd = Psr
//            1: MSR{cond} Psr{_field},Op  ;Psr[field] = Op
//   20     Must be 0b for this instruction (otherwise TST,TEQ,CMP,CMN)
//   For MRS:
//     19-16   Must be 1111b for this instruction (otherwise SWP)
//     15-12   Rd - Destination Register  (R0-R14)
//     11-0    Not used, must be zero.
//   For MSR:
//     19      f  write to flags field     Bit 31-24 (aka _flg)
//     18      s  write to status field    Bit 23-16 (reserved, don't change)
//     17      x  write to extension field Bit 15-8  (reserved, don't change)
//     16      c  write to control field   Bit 7-0   (aka _ctl)
//     15-12   Not used, must be 1111b.
//   For MSR Psr,Rm (I=0)
//     11-4    Not used, must be zero. (otherwise BX)
//     3-0     Rm - Source Register <op>  (R0-R14)
//   For MSR Psr,Imm (I=1)
//     11-8    Shift applied to Imm   (ROR in steps of two 0-30)
//     7-0     Imm - Unsigned 8bit Immediate
//     In source code, a 32bit immediate should be specified as operand.
//     The assembler should then convert that into a shifted 8bit value.

func isPSRTransfer(opcode uint32) bool {
	// MRS || MSR
	return isMatch(opcode, 0b0000_1101_1011_1111_1111_1111_1111_1111, 0b0000_0001_0000_1111_0000_0000_0000_0000) || isMatch(opcode, 0b0000_1101_0011_0000_1111_1111_1111_1111, 0b0000_0001_0010_0000_1111_1111_1111_1111)
}

//Memory: Halfword, Doubleword, and Signed Data Transfer
//   Bit    Expl.
//   31-28  Condition
//   27-25  Must be 000b for this instruction
//   24     P - Pre/Post (0=post; add offset after transfer, 1=pre; before trans.)
//   23     U - Up/Down Bit (0=down; subtract offset from base, 1=up; add to base)
//   22     I - Immediate Offset Flag (0=Register Offset, 1=Immediate Offset)
//   When above Bit 24 P=0 (Post-indexing, write-back is ALWAYS enabled):
//     21     Not used, must be zero (0)
//   When above Bit 24 P=1 (Pre-indexing, write-back is optional):
//     21     W - Write-back bit (0=no write-back, 1=write address into base)
//   20     L - Load/Store bit (0=Store to memory, 1=Load from memory)
//   19-16  Rn - Base register                (R0-R15) (Including R15=PC+8)
//   15-12  Rd - Source/Destination Register  (R0-R15) (Including R15=PC+12)
//   11-8   When above Bit 22 I=0 (Register as Offset):
//            Not used. Must be 0000b
//          When above Bit 22 I=1 (immediate as Offset):
//            Immediate Offset (upper 4bits)
//   7      Reserved, must be set (1)
//   6-5    Opcode (0-3)
//          When Bit 20 L=0 (Store) (and Doubleword Load/Store):
//           0: Reserved for SWP instruction
//           1: STR{cond}H  Rd,<Address>  ;Store halfword   [a]=Rd
//           2: LDR{cond}D  Rd,<Address>  ;Load Doubleword  R(d)=[a], R(d+1)=[a+4]
//           3: STR{cond}D  Rd,<Address>  ;Store Doubleword [a]=R(d), [a+4]=R(d+1)
//          When Bit 20 L=1 (Load):
//           0: Reserved.
//           1: LDR{cond}H  Rd,<Address>  ;Load Unsigned halfword (zero-extended)
//           2: LDR{cond}SB Rd,<Address>  ;Load Signed byte (sign extended)
//           3: LDR{cond}SH Rd,<Address>  ;Load Signed halfword (sign extended)
//   4      Reserved, must be set (1)
//   3-0    When above Bit 22 I=0:
//            Rm - Offset Register            (R0-R14) (not including R15)
//          When above Bit 22 I=1:
//            Immediate Offset (lower 4bits)  (0-255, together with upper bits)

//Double word is not supported by ARM7TDMI

func isHalfWord(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1110_0000_0000_0000_0000_1001_0000, 0b0000_00000_0000_0000_0000_0000_1001_0000)
}

// for signed data transfer, we have 2 instructions LDRSB and LDRSH
// so bit 20 should be set to 1 (store op) and 6-5 should be either 2 or 3 (10 or 11)
// 4 and 7 are always 1s
// 27-25 are always 0s
func isSignedDataTransfer(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1110_0001_0000_0000_0000_1101_0000, 0b0000_0000_0001_0000_0000_0000_1101_0000)

}

// Bit    Expl.
// 31-28  Condition
// 27-25  Must be 100b for this instruction
// 24     P - Pre/Post (0=post; add offset after transfer, 1=pre; before trans.)
// 23     U - Up/Down Bit (0=down; subtract offset from base, 1=up; add to base)
// 22     S - PSR & force user bit (0=No, 1=load PSR or force user mode)
// 21     W - Write-back bit (0=no write-back, 1=write address into base)
// 20     L - Load/Store bit (0=Store to memory, 1=Load from memory)
//
//	0: STM{cond}{amod} Rn{!},<Rlist>{^}  ;Store (Push)
//	1: LDM{cond}{amod} Rn{!},<Rlist>{^}  ;Load  (Pop)
//	Whereas, {!}=Write-Back (W), and {^}=PSR/User Mode (S)
//
// 19-16  Rn - Base register                (R0-R14) (not including R15)
// 15-0   Rlist - Register List
// (Above 'offset' is meant to be the number of words specified in Rlist.)

func isBlockDataTransfer(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1110_0000_0000_0000_0000_00000_0000, 0b0000_1000_0000_0000_0000_0000_0000_0000)
}

//  Bit    Expl.
//   31-28  Condition
//   27-23  Must be 00010b for this instruction
//          Opcode (fixed)
//            SWP{cond}{B} Rd,Rm,[Rn]      ;Rd=[Rn], [Rn]=Rm
//   22     B - Byte/Word bit (0=swap 32bit/word, 1=swap 8bit/byte)
//   21-20  Must be 00b for this instruction
//   19-16  Rn - Base register                     (R0-R14)
//   15-12  Rd - Destination Register              (R0-R14)
//   11-4   Must be 00001001b for this instruction
//   3-0    Rm - Source Register                   (R0-R14)

func isSingleDataSwap(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1111_1011_0000_0000_1111_1111_0000, 0b0000_0001_0000_0000_0000_0000_1001_0000)
}

//   Bit    Expl.
//   31-28  Condition
//   27-26  Must be 00b for this instruction
//   25     I - Immediate 2nd Operand Flag (0=Register, 1=Immediate)
//   24-21  Opcode (0-Fh)               ;*=Arithmetic, otherwise Logical
//            0: AND{cond}{S} Rd,Rn,Op2    ;AND logical       Rd = Rn AND Op2
//            1: EOR{cond}{S} Rd,Rn,Op2    ;XOR logical       Rd = Rn XOR Op2
//            2: SUB{cond}{S} Rd,Rn,Op2 ;* ;subtract          Rd = Rn-Op2
//            3: RSB{cond}{S} Rd,Rn,Op2 ;* ;subtract reversed Rd = Op2-Rn
//            4: ADD{cond}{S} Rd,Rn,Op2 ;* ;add               Rd = Rn+Op2
//            5: ADC{cond}{S} Rd,Rn,Op2 ;* ;add with carry    Rd = Rn+Op2+Cy
//            6: SBC{cond}{S} Rd,Rn,Op2 ;* ;sub with carry    Rd = Rn-Op2+Cy-1
//            7: RSC{cond}{S} Rd,Rn,Op2 ;* ;sub cy. reversed  Rd = Op2-Rn+Cy-1
//            8: TST{cond}{P}    Rn,Op2    ;test            Void = Rn AND Op2
//            9: TEQ{cond}{P}    Rn,Op2    ;test exclusive  Void = Rn XOR Op2
//            A: CMP{cond}{P}    Rn,Op2 ;* ;compare         Void = Rn-Op2
//            B: CMN{cond}{P}    Rn,Op2 ;* ;compare neg.    Void = Rn+Op2
//            C: ORR{cond}{S} Rd,Rn,Op2    ;OR logical        Rd = Rn OR Op2
//            D: MOV{cond}{S} Rd,Op2       ;move              Rd = Op2
//            E: BIC{cond}{S} Rd,Rn,Op2    ;bit clear         Rd = Rn AND NOT Op2
//            F: MVN{cond}{S} Rd,Op2       ;not               Rd = NOT Op2
//   20     S - Set Condition Codes (0=No, 1=Yes) (Must be 1 for opcode 8-B)
//   19-16  Rn - 1st Operand Register (R0..R15) (including PC=R15)
//               Must be 0000b for MOV/MVN.
//   15-12  Rd - Destination Register (R0..R15) (including PC=R15)
//               Must be 0000b (or 1111b) for CMP/CMN/TST/TEQ{P}.
//   When above Bit 25 I=0 (Register as 2nd Operand)
//     When below Bit 4 R=0 - Shift by Immediate
//       11-7   Is - Shift amount   (1-31, 0=Special/See below)
//     When below Bit 4 R=1 - Shift by Register
//       11-8   Rs - Shift register (R0-R14) - only lower 8bit 0-255 used
//       7      Reserved, must be zero  (otherwise multiply or LDREX or undefined)
//     6-5    Shift Type (0=LSL, 1=LSR, 2=ASR, 3=ROR)
//     4      R - Shift by Register Flag (0=Immediate, 1=Register)
//     3-0    Rm - 2nd Operand Register (R0..R15) (including PC=R15)
//   When above Bit 25 I=1 (Immediate as 2nd Operand)
//     11-8   Is - ROR-Shift applied to nn (0-30, in steps o

func isDataProcessing(opcode uint32) bool {
	return isMatch(opcode, 0b0000_1100_0000_0000_0000_0000_0000_0000, 0b0000_0000_0000_0000_0000_0000_0000_0000)
}
