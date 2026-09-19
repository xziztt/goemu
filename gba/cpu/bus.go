package cpu

type Bus interface {
	Read(addr uint32, size int) (uint32, error)
	Write(addr uint32, size int, value uint32) error
}

func (b *bus) Read(addr uint32, size int) (uint32, error) {
	// GBA has different memory regions, so we need to check which region the address falls into and read from the appropriate memory.

	m := addr >> 24 // Get the upper 8 bits to determine the memory region
	memoryRegion := b.GetMemoryRegion(m)
	//ToDo: handle invalid cases above

	
}

func (b *bus) GetMemoryRegion(m uint32) MemoryRegions {
	switch m {
		case m > = 0x00 && m < 0x02:
			return BiosRegion
		case m >= 0x02 && m < 0x03:
			return EWRAMRegion
		case m >= 0x03 && m < 0x04:
			return IWRAMRegion
		case m >= 0x04 && m < 0x05:
			return IORegion
		case m >= 0x05 && m < 0x06:
			return PaletteRAMRegion
		case m >= 0x06 && m < 0x07:
			return VRAMRegion
		case m >= 0x07 && m < 0x08:
			return OAMRegion
		case m >= 0x08 && m < 0x0E:
			return GamePakROM0Region
		default:
			return 0 // Invalid region
	}
}

// Each memory region has different access times, calculate cycle count based on region and size
func (b *bus) CalculateCycleCount(memoryRegion MemoryRegions, size int) int {
	
	switch memoryRegion {
		case BiosRegion:
			return 1
		case EWRAMRegion:
			return 1
		case IWRAMRegion:
			if slices.Contains([]int{8,16}, size) {
				return 3
			} else {
				return 6
			}
		case IORegion:
			return 1
		case OAMRegion:
			return 1
		case PaletteRAMRegion:
			if slices.Contains([]int{8,16}, size) {
				return 1
			} else {
				return 2
			}
		case VRAMRegion:
			if slices.Contains([]int{8,16}, size) {
				return 1
			} else {
				return 2
			}
		case GamePakROM0Region, GamePakROM1Region, GamePakROM2Region:
			if slices.Contains([]int{8,16}, size) {
				return 5
			} else {
				return 8
			}
		case GamePakSaveRegion:
			return 5
	}
}