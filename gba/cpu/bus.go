package cpu

import (
	"fmt"
	"slices"
)

type Bus interface {
	Read(addr uint32, size int) (uint32, error)
	Write(addr uint32, size int, value uint32) error
}

func GetMemoryRegion(m uint32) (MemoryRegions, error) {
	switch {
	case m >= 0x00 && m < 0x02:
		return BiosRegion, nil
	case m >= 0x02 && m < 0x03:
		return EWRAMRegion, nil
	case m >= 0x03 && m < 0x04:
		return IWRAMRegion, nil
	case m >= 0x04 && m < 0x05:
		return IORegion, nil
	case m >= 0x05 && m < 0x06:
		return PaletteRAMRegion, nil
	case m >= 0x06 && m < 0x07:
		return VRAMRegion, nil
	case m >= 0x07 && m < 0x08:
		return OAMRegion, nil
	case m >= 0x08 && m < 0x0E:
		return GamePakROM0Region, nil
	default:
		return 0, fmt.Errorf("invalid memory region") // Invalid region
	}
}

// Each memory region has different access times, calculate cycle count based on region and size
func CalculateCycleCount(memoryRegion MemoryRegions, size int) (int, error) {

	switch memoryRegion {
	case BiosRegion:
		return 1, nil
	case EWRAMRegion:
		return 1, nil
	case IWRAMRegion:
		if slices.Contains([]int{8, 16}, size) {
			return 3, nil
		} else {
			return 6, nil
		}
	case IORegion:
		return 1, nil
	case OAMRegion:
		return 1, nil
	case PaletteRAMRegion:
		if slices.Contains([]int{8, 16}, size) {
			return 1, nil
		} else {
			return 2, nil
		}
	case VRAMRegion:
		if slices.Contains([]int{8, 16}, size) {
			return 1, nil
		} else {
			return 2, nil
		}
	case GamePakROM0Region, GamePakROM1Region, GamePakROM2Region:
		if slices.Contains([]int{8, 16}, size) {
			return 5, nil
		} else {
			return 8, nil
		}
	case GamePakSaveRegion:
		return 5, nil
	}
	return 0, nil // Invalid region
}
