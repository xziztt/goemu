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
}

type CommonRegisters [16]uint32
type FIQRegisters [7]uint32
type SVCRegisters [2]uint32
type ABTRegisters [2]uint32
type IRQRegisters [2]uint32
type UNDRegisters [2]uint32

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
