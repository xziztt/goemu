package arm

// Arm opcodes are represented as 32 bit values
// This function maps the opcode to the corresponding arm operation
// https://www.akkit.org/info/gbatek.htm#arminstructionset
// Each instruction type matches a 32 bit bitmap according to the arm instruction set

func convertOpcodeToOperation(opcode uint32) func(opcode uint32) {
	switch {
	//ToDo: Implement methods to process these operations
	default:
		return nil
	}
	return nil
}
