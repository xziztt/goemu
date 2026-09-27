package cpu

type CPU struct {
	memory    MemoryRegions
	Registers Registers
}

func (c *CPU) step() {

}

// This function should be called when an operation that could possibly switch mode to THUMB occurs
func (c *CPU) switchModes() {

	if setThumb := c.Registers.Common[PC] & 0x1; setThumb == 1 {
		c.Registers.CPSR.T = true

		// Set the 0th bit to 0, PC already contains the value of Rn by this point
		// &^1 creates 1111.....0
		c.Registers.Common[PC] &^= 1
	}

	// Go back to ARM Mode
	// clearn bits 0 and 1 since ARM instruction starting addresses need to be word aligned (divisible by 4)
	c.Registers.Common[PC] &^= 3
}

// Placeholder function to skip cycles
func (c *CPU) skip(cycles uint32) {
}
