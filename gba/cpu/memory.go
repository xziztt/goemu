package cpu

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
	GamePakROMRegion  MemoryRegions = 0x08000000
	GamePakSaveRegion MemoryRegions = 0x0E000000
)

type Memory struct {
	Bios              []byte
	EWRam             []byte
	IWRam             []byte
	IORegion          []byte
	PaletteRAMRegion  []byte
	VRAMRegion        []byte
	OAMRegion         []byte
	GamePakROMRegion  []byte
	GamePakSaveRegion []byte
}
