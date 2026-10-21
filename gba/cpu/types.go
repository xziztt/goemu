// Types to be used to represent a ARM7TDMI CPU.
package cpu

type States int

const (
	ARM States = iota
	THUMB
)

type CPUModes int

const (
	User   CPUModes = iota // Normal application level mode
	System                 // Privileged application level mode
	FIQ                    // Fast interrupt request mode, handles important and time-critical interrupts
	SVC                    // Supervisor mode, handles syscalls if required
	ABT                    // Memory access exception mode
	IRQ                    // Interrupt request mode
	UND                    // Undefined instruction mode
)

type Registers struct {
	Common CommonRegisters
	FIQ    FIQRegisters
	SVC    SVCRegisters
	ABT    ABTRegisters
	IRQ    IRQRegisters
	UND    UNDRegisters
	CPSR   CPSR
	SPSR   SPSR
}

const (
	PC = 15
	LR = 14
	SP = 13
)

type (
	CommonRegisters [16]uint32
	FIQRegisters    [7]uint32
	SVCRegisters    [2]uint32
	ABTRegisters    [2]uint32
	IRQRegisters    [2]uint32
	UNDRegisters    [2]uint32
	CPSR            uint32
	SPSR            [5]uint32
)

const (
	SPSR_fiq = iota
	SPSR_svc
	SPSR_abt
	SPSR_irq
	SPSR_und
)

func (r *Registers) initialize() {
	*r = Registers{
		Common: CommonRegisters{},
		FIQ:    FIQRegisters{},
		SVC:    SVCRegisters{},
		ABT:    ABTRegisters{},
		IRQ:    IRQRegisters{},
		UND:    UNDRegisters{},
	}
}

type CPSRIndex uint32

const (
	Mode       CPSRIndex = 0
	THUMB_CPSR CPSRIndex = 5
	FIQ_CPSR   CPSRIndex = 6
	IRQ_CPSR   CPSRIndex = 7
	V          CPSRIndex = 28
	C          CPSRIndex = 29
	Z          CPSRIndex = 30
	N          CPSRIndex = 31
)

func (r *Registers) GetSPSRLoc() uint32 {
	mode := r.CPSR & 0x1F
	switch mode {
	case 0x11:
		return SPSR_fiq
	case 0x12:
		return SPSR_svc
	case 0x13:
		return SPSR_abt
	case 0x14:
		return SPSR_irq
	case 0x15:
		return SPSR_und
	}
	panic("unedfined")
}
