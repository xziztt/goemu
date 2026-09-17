package bios

type BIOS struct {
	Bios [16 * 1024]byte // load in the 16kb GBA bios
	Ptr  uint32          // pointer to the current instruction in the bios
}
