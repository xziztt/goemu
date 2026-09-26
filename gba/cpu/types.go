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
	CPSR   CPSRegisters
	SPSR   SPSRegisters
}

const (
	PC = 15
	LR = 14
	SP = 13
)

type StatusRegister struct {
	N, Z, C, V, Q, J, E, A, I, F, T bool
}

type CommonRegisters [16]uint32
type FIQRegisters [7]uint32
type SVCRegisters [2]uint32
type ABTRegisters [2]uint32
type IRQRegisters [2]uint32
type UNDRegisters [2]uint32
type CPSRegisters StatusRegister

// SPSR is used in case an IRQ occurs.
// It saves the current state of the CPSR into SPSR and switches over.
// For example, CPU is in THUMB mode and IRQ occurs, it needs to switch back to THUMB after
type SPSRegisters [6]StatusRegister

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

// Memory regions available according to https://rust-console.github.io/gbatek-gbaonly/#gbamemorymap
type MemoryRegions uint32

const (
	BiosRegion MemoryRegions = 0x00000000
	// External work RAM
	EWRAMRegion MemoryRegions = 0x02000000
	// Internal work RAM
	IWRAMRegion       MemoryRegions = 0x03000000
	IORegion          MemoryRegions = 0x04000000
	PaletteRAMRegion  MemoryRegions = 0x05000000
	VRAMRegion        MemoryRegions = 0x06000000
	OAMRegion         MemoryRegions = 0x07000000
	GamePakROM0Region MemoryRegions = 0x08000000
	GamePakROM1Region MemoryRegions = 0x0A000000
	GamePakROM2Region MemoryRegions = 0x0C000000
	GamePakSaveRegion MemoryRegions = 0x0E000000
)
